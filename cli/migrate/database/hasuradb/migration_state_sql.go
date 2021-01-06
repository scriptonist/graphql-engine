package hasuradb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hasura/graphql-engine/cli/internal/client"
	"github.com/hasura/graphql-engine/cli/migrate/database"
)

// until version 1.4 migration state was stored a special table
// this struct will implement the methods required
type migrationStateWithSQL struct {
	hasuraDB *HasuraDB
}

func NewMigrationStateStoreWithSQL(hasuraDB *HasuraDB) *migrationStateWithSQL {
	return &migrationStateWithSQL{
		hasuraDB: hasuraDB,
	}
}
func (m *migrationStateWithSQL) InsertVersion(version int64) error {
	query := HasuraQuery{
		Type: "run_sql",
		Args: HasuraArgs{
			SQL: `INSERT INTO ` + fmt.Sprintf("%s.%s", DefaultSchema, m.hasuraDB.config.MigrationsTable) + ` (version, dirty) VALUES (` + strconv.FormatInt(version, 10) + `, ` + fmt.Sprintf("%t", false) + `)`,
		},
	}
	m.hasuraDB.migrationQuery.Args = append(m.hasuraDB.migrationQuery.Args, query)
	return nil
}

func (m *migrationStateWithSQL) SetVersion(version int64, dirty bool) error {
	if version >= 0 || (version == database.NilVersion && dirty) {
		query := HasuraQuery{
			Type: "run_sql",
			Args: HasuraArgs{
				SQL: `INSERT INTO ` + fmt.Sprintf("%s.%s", DefaultSchema, m.hasuraDB.config.MigrationsTable) + ` (version, dirty) VALUES (` + strconv.FormatInt(version, 10) + `, ` + fmt.Sprintf("%t", dirty) + `)` + fmt.Sprintf(` ON DUPLICATE KEY UPDATE dirty=%t`, dirty),
			},
		}
		m.hasuraDB.migrationQuery.Args = append(m.hasuraDB.migrationQuery.Args, query)
	}
	return nil
}

func (m *migrationStateWithSQL) RemoveVersion(version int64) error {
	query := HasuraQuery{
		Type: "run_sql",
		Args: HasuraArgs{
			SQL: `DELETE FROM ` + fmt.Sprintf("%s.%s", DefaultSchema, m.hasuraDB.config.MigrationsTable) + ` WHERE version = ` + strconv.FormatInt(version, 10),
		},
	}
	m.hasuraDB.migrationQuery.Args = append(m.hasuraDB.migrationQuery.Args, query)
	return nil
}

func (m *migrationStateWithSQL) PrepareMigrationsStateStore() error {
	// check if migration table exists
	query := HasuraQuery{
		Type: "run_sql",
		Args: HasuraArgs{
			SQL: `SELECT COUNT(1) FROM information_schema.tables WHERE table_name = '` + m.hasuraDB.config.MigrationsTable + `' AND table_schema = '` + DefaultSchema + `' LIMIT 1`,
		},
	}

	resp, body, err := m.hasuraDB.SendMetadataOrQueryRequest(query, client.MetadataOrQueryClientFuncOpts{QueryRequestOpts: &client.QueryRequestOpts{}})
	if err != nil {
		m.hasuraDB.logger.Debug(err)
		return err
	}
	m.hasuraDB.logger.Debug("response: ", string(body))

	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, m.hasuraDB.config.isCMD)
	}

	var hres HasuraSQLRes
	err = json.Unmarshal(body, &hres)
	if err != nil {
		m.hasuraDB.logger.Debug(err)
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
			SQL: `CREATE TABLE ` + fmt.Sprintf("%s.%s", DefaultSchema, m.hasuraDB.config.MigrationsTable) + ` (version bigint not null primary key, dirty boolean not null)`,
		},
	}

	resp, body, err = m.hasuraDB.SendMetadataOrQueryRequest(query, client.MetadataOrQueryClientFuncOpts{QueryRequestOpts: &client.QueryRequestOpts{}})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, m.hasuraDB.config.isCMD)
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

func (m *migrationStateWithSQL) GetVersions() error {
	query := HasuraQuery{
		Type: "run_sql",
		Args: HasuraArgs{
			SQL: `SELECT version, dirty FROM ` + fmt.Sprintf("%s.%s", DefaultSchema, m.hasuraDB.config.MigrationsTable),
		},
	}

	// Send Query
	resp, body, err := m.hasuraDB.SendMetadataOrQueryRequest(query, client.MetadataOrQueryClientFuncOpts{QueryRequestOpts: &client.QueryRequestOpts{}})
	if err != nil {
		return err
	}

	// If status != 200 return error
	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, m.hasuraDB.config.isCMD)
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

		m.hasuraDB.migrations.Append(version)
	}

	return nil
}
