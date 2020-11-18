package hasuradb

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const (
	DefaultSettingsTable = "migration_settings"
)

func (h *HasuraDB) ensureSettingsTable(hasDataSources bool) error {
	// check if migration table exists
	if !hasDataSources {
		query := HasuraQuery{
			Type: "run_sql",
			Args: HasuraArgs{
				SQL: `SELECT COUNT(1) FROM information_schema.tables WHERE table_name = '` + h.config.SettingsTable + `' AND table_schema = '` + DefaultSchema + `' LIMIT 1`,
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
				SQL: `CREATE TABLE ` + fmt.Sprintf("%s.%s", DefaultSchema, h.config.SettingsTable) + ` (setting text not null primary key, value text not null)`,
			},
		}

		resp, body, err = h.sendv1Query(query)
		if err != nil {
			return err
		}
		h.logger.Debug("response: ", string(body))

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
		return h.setDefaultSettings(false)
	}

	resp, body, err := h.sendV1MetadataQuery(GetCatalogState, map[string]interface{}{}, "")

	if err != nil {
		return errors.Wrap(err, "Could not fetch the latest state from the database")
	}

	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, h.config.isCMD)
	}

	var currentCatalogState map[string]interface{}
	if err := json.Unmarshal(body, &currentCatalogState); err != nil {
		// FIXME: probably should be handling this better
		return err
	}

	toSetKey := false

	if currentCatalogState[DefaultMigrationsTable] == nil {
		toSetKey = true
	}

	if toSetKey {
		args := map[string]interface{}{
			DefaultMigrationsTable: map[string]interface{}{},
		}

		setResp, setBody, setErr := h.sendV1MetadataQuery(SetCatalogState, args, "")

		if setErr != nil {
			return errors.Wrap(err, "Failed to set the settings key on the cli state")
		}

		if setResp.StatusCode != http.StatusOK {
			return NewHasuraError(setBody, h.config.isCMD)
		}
	}

	return h.setDefaultSettings(true)
}

func (h *HasuraDB) setDefaultSettings(hasDataSources bool) error {
	if !hasDataSources {
		query := HasuraBulk{
			Type: "bulk",
			Args: make([]HasuraQuery, 0),
		}
		for _, setting := range h.settings {
			sql := HasuraQuery{
				Type: "run_sql",
				Args: HasuraArgs{
					SQL: `INSERT INTO ` + fmt.Sprintf("%s.%s", DefaultSchema, h.config.SettingsTable) + ` (setting, value) VALUES ('` + fmt.Sprintf("%s", setting.GetName()) + `', '` + fmt.Sprintf("%s", setting.GetDefaultValue()) + `')`,
				},
			}
			query.Args = append(query.Args, sql)
		}

		if len(query.Args) == 0 {
			return nil
		}

		resp, body, err := h.sendv1Query(query)
		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			return NewHasuraError(body, h.config.isCMD)
		}
	}

	args := map[string]interface{}{
		DefaultMigrationsTable: true,
	}

	resp, body, err := h.sendV1MetadataQuery(SetCatalogState, args, "")

	if err != nil {
		return errors.Wrap(err, "Could not update CLI state correctly")
	}

	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, h.config.isCMD)
	}

	return nil
}

// TODO: have to change this for data sources
func (h *HasuraDB) GetSetting(name string) (value string, err error) {
	query := HasuraQuery{
		Type: "run_sql",
		Args: HasuraArgs{
			SQL: `SELECT value from ` + fmt.Sprintf("%s.%s", DefaultSchema, h.config.SettingsTable) + ` where setting='` + name + `'`,
		},
	}

	// Send Query
	resp, body, err := h.sendv1Query(query)
	if err != nil {
		return value, err
	}
	h.logger.Debug("response: ", string(body))

	// If status != 200 return error
	if resp.StatusCode != http.StatusOK {
		return value, NewHasuraError(body, h.config.isCMD)
	}

	var hres HasuraSQLRes
	err = json.Unmarshal(body, &hres)
	if err != nil {
		return value, err
	}

	if hres.ResultType != TuplesOK {
		return value, fmt.Errorf("Invalid result Type %s", hres.ResultType)
	}

	if len(hres.Result) < 2 {
		for _, setting := range h.settings {
			if setting.GetName() == name {
				return setting.GetDefaultValue(), nil
			}
		}
		return value, fmt.Errorf("Invalid setting name: %s", name)
	}

	return hres.Result[1][0], nil
}

// TODO: have to change this for data sources
func (h *HasuraDB) UpdateSetting(name string, value string) error {
	query := HasuraQuery{
		Type: "run_sql",
		Args: HasuraArgs{
			SQL: `INSERT INTO ` + fmt.Sprintf("%s.%s", DefaultSchema, h.config.SettingsTable) + ` (setting, value) VALUES ('` + name + `', '` + value + `') ON CONFLICT (setting) DO UPDATE SET value='` + value + `'`,
		},
	}

	// Send Query
	resp, body, err := h.sendv1Query(query)
	if err != nil {
		return err
	}
	h.logger.Debug("response: ", string(body))

	// If status != 200 return error
	if resp.StatusCode != http.StatusOK {
		return NewHasuraError(body, h.config.isCMD)
	}

	var hres HasuraSQLRes
	err = json.Unmarshal(body, &hres)
	if err != nil {
		return err
	}

	if hres.ResultType != CommandOK {
		return fmt.Errorf("Cannot set setting %s to %s", name, value)
	}
	return nil
}
