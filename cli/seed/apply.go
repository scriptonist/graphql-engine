package seed

import (
	"fmt"
	"log"
	"net/http"
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

	files, err := afero.ReadDir(fs, directoryPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			b, err := afero.ReadFile(fs, filepath.Join(directoryPath, file.Name()))
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
	}

	resp, b, err := hasuraAPIProvider.SendQuery(seedQuery)
	if err != nil {
		return errors.Wrap(err, "error running V1 hasura query")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.Wrap(errors.New("error executing V1 hasura query"), fmt.Sprintf("%v%s", resp, string(b)))
	}

	return nil
}
