package util

import (
	"crypto/tls"

	"github.com/Masterminds/semver"
	"github.com/hasura/graphql-engine/cli/version"
	"github.com/parnurzeal/gorequest"
	"github.com/sirupsen/logrus"
)

// MigrationState is derived from the schema_migrations table
type MigrationState struct {
	Version int64  `json:"version"`
	Name    string `json:"name"` // may not be necessary
	Dirty   bool   `json:"dirty"`
}

// CLIState is derived from the migration_settings and cli_state
type CLIState struct {
	Migrations    []MigrationState `json:"migrations"`
	MigrationMode bool             `json:"migration_mode"`
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

type hdbVersion struct {
	UUID     string                 `json:"hasura_uuid"`
	CLIState map[string]interface{} `json:"cli_state"`
}

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
		state.CLIState = r[0].CLIState
		return state
	}

	var res hdbVersionV2
	_, _, errs := req.EndStruct(&res)

	if len(errs) != 0 {
		log.Debugf("server state: errors: %v", errs)
		return state
	}

	state.UUID = res.ID
	// FIXME: definitely sure that this is not ideal, but this is one way to get around the current constraints
	state.CLIState = map[string]interface{}{
		"state": res.CLIState,
	}
	return state
}

// UpdateServerState will update the server state with a new state object
func UpdateServerState(serverFeatureFlags *version.ServerFeatureFlags, updatedState CLIState) error {
	// TODO: complete the implementation
	// 1. check if all the required fields are present (everything that is part of the CLIState)
	// 2. make the request body
	/*{
			type: set_catalog_state
			args: {
			  "type": "console",
	    	  "state": {"key": "value"}
			}
		  }
	*/
	return nil
}
