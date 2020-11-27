package hasuradb

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"io/ioutil"
	"net/http"
	nurl "net/url"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/hasura/graphql-engine/cli/version"

	yaml "github.com/ghodss/yaml"
	"github.com/hasura/graphql-engine/cli/metadata/types"
	"github.com/hasura/graphql-engine/cli/migrate/database"
	"github.com/oliveagle/jsonpath"
	"github.com/parnurzeal/gorequest"
	log "github.com/sirupsen/logrus"
)

func init() {
	db := HasuraDB{}
	database.Register("hasuradb", &db)
}

// TODO: for data sources, there won't be a migrations table, but instead it will have a key on the CLI state on the DB
const (
	DefaultMigrationsTable = "schema_migrations"
	DefaultSchema          = "hdb_catalog"
)

var (
	ErrNilConfig      = fmt.Errorf("no config")
	ErrNoDatabaseName = fmt.Errorf("no database name")
	ErrNoSchema       = fmt.Errorf("no schema")
	ErrDatabaseDirty  = fmt.Errorf("database is dirty")
)

type Config struct {
	MigrationsTable                string
	SettingsTable                  string
	queryURL                       *nurl.URL
	graphqlURL                     *nurl.URL
	pgDumpURL                      *nurl.URL
	metadataURL                    *nurl.URL
	Headers                        map[string]string
	isCMD                          bool
	Plugins                        types.MetadataPlugins
	enableCheckMetadataConsistency bool
	Req                            *gorequest.SuperAgent
}

type DataSourceURL struct {
	FromEnv   string `json:"from_env,omitempty"`
	FromValue string `json:"from_value,omitempty"`
}

type ConnectionPoolSettings struct {
	MaxConnections        int32 `json:"max_connections,omitempty"`
	ConnectionIdleTimeout int32 `json:"connection_idle_timeout,omitempty"`
}

type QualifiedTable struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type TableEntry struct {
	Table  QualifiedTable `json:"table"`
	IsEnum bool           `json:"is_enum,omitempty"`
}

type ActionPermissions struct {
	Role string `json:"role"`
}

type InputArgument struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ServerHeader struct {
	Name         string `json:"name"`
	Value        string `json:"value,omitempty"`
	ValueFromEnv string `json:"value_from_env,omitempty"`
}

type ActionDefinition struct {
	Handler              string          `json:"handler"`
	Type                 string          `json:"type,omitempty"` // Can be either mutation / query
	ForwardClientHeaders bool            `json:"forward_client_headers,omitempty"`
	Kind                 string          `json:"kind,omitempty"` // Can be either synchronous / asynchronous
	OutputType           string          `json:"output_type,omitempty"`
	Arguments            []InputArgument `json:"arguments,omitempty"`
	Headers              []ServerHeader  `json:"headers,omitempty"`
}

type Action struct {
	Name        string              `json:"name"`
	Comment     string              `json:"comment,omitempty"`
	Permissions []ActionPermissions `json:"permissions,omitempty"`
	Definition  ActionDefinition    `json:"definition"`
}

type V3Metadata struct {
	Name                   string                 `json:"name"`
	URL                    DataSourceURL          `json:"url"`
	ConnectionPoolSettings ConnectionPoolSettings `json:"connection_pool_settings,omitempty"`
	Tables                 []TableEntry           `json:"tables"`
	Actions                []Action               `json:"actions,omitempty"`
	//CustomTypes CustomTypes `json:"custom_types,omitempty"` -- Don't think this is required at the moment
}

type TotalMetadata struct {
	Version int          `json:"version"`
	Sources []V3Metadata `json:"sources"`
}

type Migration struct {
	Name    string `json:"name,omitempty"`
	Dirty   bool   `json:"dirty,omitempty"`
	Version int64  `json:"version,omitempty"`
}

type MigrationSetting struct {
	MigrationMode bool `json:"migration_mode,omitempty"`
}

type CLIState struct {
	SchemaMigrations  []Migration      `json:"schema_migrations,omitempty"`
	MigrationsSetting MigrationSetting `json:"migration_settings,omitempty"`
}

type CatalogState struct {
	ID       string   `json:"id"`
	CLIState CLIState `json:"cli_state,omitempty"`
}

type HasuraDB struct {
	config             *Config
	settings           []database.Setting
	migrations         *database.Migrations
	migrationQuery     HasuraInterfaceBulk
	jsonPath           map[string]string
	isLocked           bool
	logger             *log.Logger
	serverFeatureFlags version.ServerFeatureFlags
	currentSource      string
	currentMetadata    TotalMetadata
	connectedSources   []string
	catalogState       CatalogState
}

func WithInstance(config *Config, logger *log.Logger, serverFeatureFlags version.ServerFeatureFlags) (database.Driver, error) {
	if config == nil {
		logger.Debug(ErrNilConfig)
		return nil, ErrNilConfig
	}

	hx := &HasuraDB{
		config:             config,
		migrations:         database.NewMigrations(),
		settings:           database.Settings,
		logger:             logger,
		serverFeatureFlags: serverFeatureFlags,
		currentSource:      "default",
	}

	if err := hx.ensureVersionTable(); err != nil {
		logger.Debug(err)
		return nil, err
	}

	if err := hx.ensureSettingsTable(); err != nil {
		logger.Debug(err)
		return nil, err
	}

	return hx, nil
}

func (h *HasuraDB) Open(url string, isCMD bool, tlsConfig *tls.Config, logger *log.Logger, serverFeatureFlags version.ServerFeatureFlags) (database.Driver, error) {
	if logger == nil {
		logger = log.New()
	}
	hurl, err := nurl.Parse(url)
	if err != nil {
		logger.Debug(err)
		return nil, err
	}
	// Use sslMode query param to set Scheme
	var scheme string
	params := hurl.Query()
	sslMode := params.Get("sslmode")
	if sslMode == "enable" {
		scheme = "https"
	} else {
		scheme = "http"
	}

	headers := make(map[string]string)
	if queryHeaders, ok := params["headers"]; ok {
		for _, header := range queryHeaders {
			headerValue := strings.SplitN(header, ":", 2)
			if len(headerValue) == 2 && headerValue[1] != "" {
				headers[headerValue[0]] = headerValue[1]
			}
		}
	}

	req := gorequest.New()
	if tlsConfig != nil {
		req.TLSClientConfig(tlsConfig)
	}

	config := &Config{
		MigrationsTable: DefaultMigrationsTable,
		SettingsTable:   DefaultSettingsTable,
		queryURL: &nurl.URL{
			Scheme: scheme,
			Host:   hurl.Host,
			Path:   path.Join(hurl.Path, params.Get("query")),
		},
		graphqlURL: &nurl.URL{
			Scheme: scheme,
			Host:   hurl.Host,
			Path:   path.Join(hurl.Path, params.Get("graphql")),
		},
		pgDumpURL: &nurl.URL{
			Scheme: scheme,
			Host:   hurl.Host,
			Path:   path.Join(hurl.Path, params.Get("pg_dump")),
		},
		metadataURL: &nurl.URL{
			Scheme: scheme,
			Host:   hurl.Host,
			Path:   path.Join(hurl.Path, params.Get("metadata")),
		},
		isCMD:   isCMD,
		Headers: headers,
		Plugins: make(types.MetadataPlugins, 0),
		Req:     req,
	}
	hx, err := WithInstance(config, logger, serverFeatureFlags)
	if err != nil {
		logger.Debug(err)
		return nil, err
	}
	return hx, nil
}

func (h *HasuraDB) Close() error {
	// nothing do to here
	return nil
}

func (h *HasuraDB) Scan() error {
	h.migrations = database.NewMigrations()
	return h.getVersions()
}

func (h *HasuraDB) Lock() error {
	if h.isLocked {
		return database.ErrLocked
	}

	h.migrationQuery = HasuraInterfaceBulk{
		Type: "bulk",
		Args: make([]interface{}, 0),
	}
	h.jsonPath = make(map[string]string)
	h.isLocked = true
	return nil
}

func (h *HasuraDB) UnLock() error {
	if !h.isLocked {
		return nil
	}

	defer func() {
		h.isLocked = false
	}()

	if len(h.migrationQuery.Args) == 0 {
		return nil
	}

	resp, body, err := h.sendv1Query(h.migrationQuery)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		switch herror := NewHasuraError(body, h.config.isCMD).(type) {
		case HasuraError:
			// Handle migration version here
			if herror.Path != "" {
				jsonData, err := json.Marshal(h.migrationQuery)
				if err != nil {
					return err
				}
				var migrationQuery interface{}
				err = json.Unmarshal(jsonData, &migrationQuery)
				if err != nil {
					return err
				}
				res, err := jsonpath.JsonPathLookup(migrationQuery, herror.Path)
				if err == nil {
					queryData, err := json.MarshalIndent(res, "", "    ")
					if err != nil {
						return err
					}
					herror.migrationQuery = string(queryData)
				}
				re1, err := regexp.Compile(`\$.args\[([0-9]+)\]*`)
				if err != nil {
					return err
				}
				result := re1.FindAllStringSubmatch(herror.Path, -1)
				if len(result) != 0 {
					migrationNumber, ok := h.jsonPath[result[0][1]]
					if ok {
						herror.migrationFile = migrationNumber
					}
				}
			}
			return herror
		default:
			return herror
		}
	}
	return nil
}

func (h *HasuraDB) Run(migration io.Reader, fileType, fileName string) error {
	migr, err := ioutil.ReadAll(migration)
	if err != nil {
		return err
	}
	body := string(migr[:])
	switch fileType {
	case "sql":
		if body == "" {
			break
		}
		sqlInput := RunSQLInput{
			SQL: string(body),
		}
		if h.config.enableCheckMetadataConsistency {
			sqlInput.CheckMetadataConsistency = func() *bool { b := false; return &b }()
		}
		t := HasuraInterfaceQuery{
			Type: RunSQL,
			Args: sqlInput,
		}
		h.migrationQuery.Args = append(h.migrationQuery.Args, t)
		h.jsonPath[fmt.Sprintf("%d", len(h.migrationQuery.Args)-1)] = fileName
	case "meta":
		var t []interface{}
		err := yaml.Unmarshal(migr, &t)
		if err != nil {
			h.migrationQuery.ResetArgs()
			return err
		}

		for _, v := range t {
			h.migrationQuery.Args = append(h.migrationQuery.Args, v)
			h.jsonPath[fmt.Sprintf("%d", len(h.migrationQuery.Args)-1)] = fileName
		}
	}
	return nil
}

func (h *HasuraDB) ResetQuery() {
	h.migrationQuery.ResetArgs()
}

func (h *HasuraDB) InsertVersion(version int64) error {
	if !h.serverFeatureFlags.HasDatasources {
		query := HasuraQuery{
			Type: "run_sql",
			Args: HasuraArgs{
				SQL: `INSERT INTO ` + fmt.Sprintf("%s.%s", DefaultSchema, h.config.MigrationsTable) + ` (version, dirty) VALUES (` + strconv.FormatInt(version, 10) + `, ` + fmt.Sprintf("%t", false) + `)`,
			},
		}
		h.migrationQuery.Args = append(h.migrationQuery.Args, query)
		return nil
	}

	migrationObj := Migration{
		Version: version,
		Dirty:   false,
	}

	// TODO: we need a place to set the cli state, and ensure that's updated all the time, we perform an API action

	h.catalogState.CLIState.SchemaMigrations = append(h.catalogState.CLIState.SchemaMigrations, migrationObj)
	data, err := json.Marshal(h.catalogState.CLIState)

	if err != nil {
		return errors.Wrap(err, err.Error())
	}

	var newState map[string]interface{}
	json.Unmarshal(data, &newState)

	args := map[string]interface{}{
		"state": newState,
		"type":  "cli",
	}

	resp, body, err := h.sendV1MetadataQuery(SetCatalogState, args, "")

	if err != nil {
		errors.Wrap(err, "Failed to update details of migration to the database")
	}

	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, h.config.isCMD)
	}

	return nil
}

func (h *HasuraDB) RemoveVersion(version int64) error {
	if !h.serverFeatureFlags.HasDatasources {
		query := HasuraQuery{
			Type: "run_sql",
			Args: HasuraArgs{
				SQL: `DELETE FROM ` + fmt.Sprintf("%s.%s", DefaultSchema, h.config.MigrationsTable) + ` WHERE version = ` + strconv.FormatInt(version, 10),
			},
		}
		h.migrationQuery.Args = append(h.migrationQuery.Args, query)
		return nil
	}

	// NOTE: This strictly assumes that the hasuradb struct contains a copy of the latest catalog state

	// remove the migration from the list
	newSetOfMigrations := make([]Migration, len(h.catalogState.CLIState.SchemaMigrations)-1)
	for _, mig := range h.catalogState.CLIState.SchemaMigrations {
		if version != mig.Version {
			newSetOfMigrations = append(newSetOfMigrations, mig)
		}
	}

	h.catalogState.CLIState.SchemaMigrations = newSetOfMigrations

	data, err := json.Marshal(h.catalogState.CLIState)

	if err != nil {
		return errors.Wrap(err, err.Error())
	}

	var newState map[string]interface{}
	json.Unmarshal(data, &newState)

	args := map[string]interface{}{
		"state": newState,
		"type":  "cli",
	}

	resp, body, err := h.sendV1MetadataQuery(SetCatalogState, args, "")

	if err != nil {
		errors.Wrap(err, "Failed to update details of migration to the database")
	}

	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, h.config.isCMD)
	}

	return nil
}

// TODO: change for data sources changes
func (h *HasuraDB) getVersions() (err error) {

	query := HasuraQuery{
		Type: "run_sql",
		Args: HasuraArgs{
			SQL: `SELECT version, dirty FROM ` + fmt.Sprintf("%s.%s", DefaultSchema, h.config.MigrationsTable),
		},
	}

	// Send Query
	resp, body, err := h.sendv1Query(query)
	if err != nil {
		return err
	}

	// If status != 200 return error
	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, h.config.isCMD)
	}

	var hres HasuraSQLRes
	err = json.Unmarshal(body, &hres)
	if err != nil {
		return err
	}

	if hres.ResultType != TuplesOK {
		return fmt.Errorf("Invalid result Type %s", hres.ResultType)
	}

	if len(hres.Result) == 1 {
		return nil
	}

	for index, val := range hres.Result {
		if index == 0 {
			continue
		}

		version, err := strconv.ParseUint(val[0], 10, 64)
		if err != nil {
			return err
		}

		h.migrations.Append(version)
	}

	return nil
}

func (h *HasuraDB) Version() (version int64, dirty bool, err error) {
	tmpVersion, ok := h.migrations.Last()
	if !ok {
		return database.NilVersion, false, nil
	}

	return int64(tmpVersion), false, nil
}

func (h *HasuraDB) Drop() error {
	return nil
}

func (h *HasuraDB) ensureVersionTable() error {
	// check if migration table exists
	if !h.serverFeatureFlags.HasDatasources {
		query := HasuraQuery{
			Type: "run_sql",
			Args: HasuraArgs{
				SQL: `SELECT COUNT(1) FROM information_schema.tables WHERE table_name = '` + h.config.MigrationsTable + `' AND table_schema = '` + DefaultSchema + `' LIMIT 1`,
			},
		}

		resp, body, err := h.sendv1Query(query)
		if err != nil {
			h.logger.Debug(err)
			return err
		}
		h.logger.Debug("response: ", string(body))

		if resp.StatusCode != http.StatusOK {
			return NewHasuraError(body, h.config.isCMD)
		}

		var hres HasuraSQLRes
		err = json.Unmarshal(body, &hres)
		if err != nil {
			h.logger.Debug(err)
			return err
		}

		if hres.ResultType != TuplesOK {
			return fmt.Errorf("Invalid result Type %s", hres.ResultType)
		}

		if hres.Result[1][0] != "0" {
			return nil
		}

		// Now Create the table
		query = HasuraQuery{
			Type: "run_sql",
			Args: HasuraArgs{
				SQL: `CREATE TABLE ` + fmt.Sprintf("%s.%s", DefaultSchema, h.config.MigrationsTable) + ` (version bigint not null primary key, dirty boolean not null)`,
			},
		}

		resp, body, err = h.sendv1Query(query)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return NewHasuraError(body, h.config.isCMD)
		}

		err = json.Unmarshal(body, &hres)
		if err != nil {
			return err
		}

		if hres.ResultType != CommandOK {
			return fmt.Errorf("Creating Version table failed %s", hres.ResultType)
		}

		return nil
	}

	resp, body, err := h.sendV1MetadataQuery(GetCatalogState, map[string]interface{}{}, "")

	var currentCatalogState map[string]interface{}
	json.Unmarshal(body, &currentCatalogState)

	if err != nil {
		return errors.Wrap(err, "Failed to fetch the latest details on the migrations")
	}

	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, h.config.isCMD)
	}

	toBeSet := false
	if currentCatalogState["cli_state"] != nil && currentCatalogState["cli_state"].(map[string]interface{})[DefaultMigrationsTable] == nil {
		toBeSet = true
	}

	if toBeSet {
		currentCliState := currentCatalogState["cli_state"]
		currentCliState.(map[string]interface{})[DefaultMigrationsTable] = map[string]interface{}{}
		args := map[string]interface{}{
			"state": currentCliState,
			"type":  "cli",
		}

		respSet, bodySet, errSet := h.sendV1MetadataQuery(SetCatalogState, args, "")

		if errSet != nil {
			return errors.Wrap(err, "Failed to set the latest state on the database")
		}

		if respSet.StatusCode != http.StatusOK {
			return NewHasuraError(bodySet, h.config.isCMD)
		}
	}

	return nil
}

func (h *HasuraDB) sendv1Query(m interface{}) (resp *http.Response, body []byte, err error) {
	request := h.config.Req.Clone()
	request = request.Post(h.config.queryURL.String()).Send(m)
	for headerName, headerValue := range h.config.Headers {
		request.Set(headerName, headerValue)
	}

	resp, body, errs := request.EndBytes()

	if len(errs) == 0 {
		err = nil
	} else {
		err = errs[0]
	}

	return resp, body, err
}

func (h *HasuraDB) sendv2Query(queryType V2Query, args map[string]interface{}, source string) (resp *http.Response, body []byte, err error) {
	requestEndpoint := h.config.queryURL.String()
	if requestEndpoint != string(V2QueryEndpoint) {
		requestEndpoint = string(V2QueryEndpoint)
	}

	payload := GetV2Query(queryType, args, source)

	request := h.config.Req.Clone()
	request = request.Post(requestEndpoint).Send(payload)
	for headerName, headerValue := range h.config.Headers {
		request.Set(headerName, headerValue)
	}

	resp, body, errs := request.EndBytes()

	if len(errs) == 0 {
		err = nil
	} else {
		err = errs[0]
	}

	return resp, body, err
}

func (h *HasuraDB) sendV1MetadataQuery(queryType V1Metadata, args map[string]interface{}, source string) (resp *http.Response, body []byte, err error) {
	requestEndpoint := h.config.metadataURL.String()
	if requestEndpoint != string(V1MetadataEndpoint) {
		requestEndpoint = string(V1MetadataEndpoint)
	}

	// FIXME : passing Postgres here since, that's what is supported at the moment
	payload := GetV1MetadataQuery(queryType, args, Postgres, source)

	request := h.config.Req.Clone()
	request = request.Post(requestEndpoint).Send(payload)
	for headerName, headerValue := range h.config.Headers {
		request.Set(headerName, headerValue)
	}

	resp, body, errs := request.EndBytes()

	if len(errs) == 0 {
		err = nil
	} else {
		err = errs[0]
	}

	return resp, body, err
}

func (h *HasuraDB) sendv1GraphQL(query interface{}) (resp *http.Response, body []byte, err error) {
	request := h.config.Req.Clone()
	request = request.Post(h.config.graphqlURL.String()).Send(query)

	for headerName, headerValue := range h.config.Headers {
		request.Set(headerName, headerValue)
	}

	resp, body, errs := request.EndBytes()

	if len(errs) == 0 {
		err = nil
	} else {
		err = errs[0]
	}

	return resp, body, err
}

func (h *HasuraDB) sendSchemaDumpQuery(m interface{}) (resp *http.Response, body []byte, err error) {
	request := h.config.Req.Clone()

	request = request.Post(h.config.pgDumpURL.String()).Send(m)

	for headerName, headerValue := range h.config.Headers {
		request.Set(headerName, headerValue)
	}

	resp, body, errs := request.EndBytes()

	if len(errs) == 0 {
		err = nil
	} else {
		err = errs[0]
	}

	return resp, body, err
}

func (h *HasuraDB) First() (version uint64, ok bool) {
	return h.migrations.First()
}

func (h *HasuraDB) Last() (version uint64, ok bool) {
	return h.migrations.Last()
}

func (h *HasuraDB) Prev(version uint64) (prevVersion uint64, ok bool) {
	return h.migrations.Prev(version)
}

func (h *HasuraDB) Next(version uint64) (nextVersion uint64, ok bool) {
	return h.migrations.Next(version)
}

func (h *HasuraDB) Read(version uint64) (ok bool) {
	return h.migrations.Read(version)
}
