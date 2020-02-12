package seed

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/hasura/graphql-engine/cli/migrate/database/hasuradb"
	"github.com/spf13/afero"
)

// HasuraAPIProvider will facilitate the interactions with hasura
type HasuraAPIProvider interface {
	SendQuery(m interface{}) (*http.Response, []byte, error)
}

// ApplySeedsToDatabase will read all .sql files in the given
// directory and apply it to hasura
func ApplySeedsToDatabase(fs afero.Fs, hasuraAPIProvider HasuraAPIProvider, directoryPath string) error {
	seedQuery := hasuradb.HasuraInterfaceBulk{
		Type: "bulk",
		Args: make([]interface{}, 0),
	}
	err := afero.Walk(fs, directoryPath, func(path string, file os.FileInfo, err error) error {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".sql" {
			b, err := afero.ReadFile(fs, path)
			if err != nil {
				return err
			}
			q := hasuradb.HasuraInterfaceQuery{
				Type: "run_sql",
				Args: hasuradb.HasuraArgs{
					SQL: string(b),
				},
			}
			seedQuery.Args = append(seedQuery.Args, q)
		}

		return nil
	})
	if err != nil {
		return err
	}

	resp, b, err := hasuraAPIProvider.SendQuery(seedQuery)
	if err != nil {
		return errors.Wrap(err, "error running hasura query")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Wrap(errors.New("error executing hasura query"), fmt.Sprintf("%v%s", resp, string(b)))
	}

	return nil
}
