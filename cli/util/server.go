package util

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/hasura/graphql-engine/cli/version"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

// MigrationState is derived from the schema_migrations table
type MigrationState struct {
	Version int64  `json:"version"`
	Name    string `json:"name,omitempty"` // may not be necessary
	Dirty   bool   `json:"dirty"`
}

// MigrationSetting are all fields that were in the migration_settings table
// NOTE: the type is (text, text)
type MigrationSetting struct {
	MigrationMode bool `json:"migration_mode"`
	// TODO: add other settings here
}

// CLIState is derived from the migration_settings and cli_state
type CLIState struct {
	Migrations       []MigrationState `json:"migrations"`
	MigrationSetting MigrationSetting `json:"migration_settings"`
	// TODO: add more fields here
}

// ServerState is the state of Hasura stored on the server.
type ServerState struct {
	// will be there in versions pre-1.4
	UUID string
	// will be there in version beyond v1.4
	ID       string
	CLIState map[string]interface{}
}

// hdbVersion will be used for the version(s) that do not support data sources
type hdbVersion struct {
	UUID     string                 `json:"hasura_uuid"`
	CLIState map[string]interface{} `json:"cli_state"`
}

// hdbVersionV2 will be used for the version(s) that support data sources
type hdbVersionV2 struct {
	ID       string   `json:"id"`
	CLIState CLIState `json:"cli_state"`
}

const defaultUUID = "00000000-0000-0000-0000-000000000000"

// GetServerState queries a server for the state.
func GetServerState(adminSecret string, config *tls.Config, serverVersion *semver.Version, log *logrus.Logger, serverFeatureFlags *version.ServerFeatureFlags) *ServerState {
	requestEndpoint := "v1/query"
	state := &ServerState{
		UUID: defaultUUID,
	}
	payload := `{
		"type": "select",
		"args": {
			"table": {
				"schema": "hdb_catalog",
				"name": "hdb_version"
			},
			"columns": [
				"hasura_uuid",
				"cli_state"
			]
		}
	}`

	if serverFeatureFlags.HasDatasources {
		requestEndpoint = "v1/metadata"
		state = &ServerState{
			ID: defaultUUID,
		}
		payload = `{
		   "type": "get_catalog_state"
		   "args": {}
		}`
	}

	req := gorequest.New()
	if config != nil {
		req.TLSClientConfig(config)
	}
	req.Post(requestEndpoint).Send(payload)
	req.Set("X-Hasura-Admin-Secret", adminSecret)

	if !serverFeatureFlags.HasDatasources {
		var r []hdbVersion
		_, _, errs := req.EndStruct(&r)
		if len(errs) != 0 {
			log.Debugf("server state: errors: %v", errs)
			return state
		}

		if len(r) != 1 {
			log.Debugf("invalid response: %v", r)
			return state
		}

		state.UUID = r[0].UUID
		state.CLIState = map[string]interface{}{
			"v1": r[0].CLIState,
		}
		return state
	}

	var res hdbVersionV2
	_, _, errs := req.EndStruct(&res)

	if len(errs) != 0 {
		log.Debugf("server state: errors: %v", errs)
		return state
	}

	state.ID = res.ID
	state.UUID = res.ID
	// FIXME: definitely sure that this is not ideal, but this is one way to get around the current constraints
	state.CLIState = map[string]interface{}{
		"v2": res.CLIState,
	}
	return state
}

// UpdateServerState will update the server state with a new state object
func UpdateServerState(serverFeatureFlags *version.ServerFeatureFlags, adminSecret string, config *tls.Config, log *logrus.Logger, updatedState CLIState) (*ServerState, error) {
	latestState := &ServerState{
		ID: defaultUUID,
	}

	if !serverFeatureFlags.HasDatasources {
		// TODO: improve the error message
		return latestState, errors.New("the current data source does not support the usage of the update state API")
	}
	// data sources are supported

	updatedStateJSON, err := json.Marshal(updatedState)

	if err != nil {
		return latestState, errors.New("Failed to update state on the DB")
	}

	payload := fmt.Sprintf(`{
		type: "set_catalog_state",
		args: {
			"type": "cli",
			"state": %s,
			}
		}`, string(updatedStateJSON))

	endpoint := "v1/metadata"
	req := gorequest.New()
	if config != nil {
		req.TLSClientConfig(config)
	}
	req.Post(endpoint).Send(payload)
	req.Set("X-Hasura-Admin-Secret", adminSecret)

	var res hdbVersionV2

	_, _, errs := req.EndStruct(&res)

	if len(errs) != 0 {
		log.Debugf("failed to update the server state: errors: %v", errs)
		// FIXME: errs[0] is not a good solution, it might just one of many.
		return latestState, errs[0]
	}

	latestState.UUID = res.ID
	latestState.ID = res.ID
	latestState.CLIState = map[string]interface{}{
		"v2": res.CLIState,
	}

	return latestState, nil
}
