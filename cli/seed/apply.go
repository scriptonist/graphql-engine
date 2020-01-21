package seed

import (
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"

	"github.com/pkg/errors"

	"github.com/hasura/graphql-engine/cli/migrate/database/hasuradb"
)

// HasuraAPIProvider will facilitate the interactions with hasura
type HasuraAPIProvider interface {
	sendv1Query(m interface{}) (*http.Response, []byte, error)
}

// ApplySeedsToDatabase will read all .sql files in the given
// directory and apply it to hasura
func ApplySeedsToDatabase(directoryPath string, provider HasuraAPIProvider) error {
	seedQuery := hasuradb.HasuraInterfaceBulk{
		Type: "bulk",
		Args: make([]interface{}, 0),
	}

	files, err := ioutil.ReadDir(directoryPath)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".sql" {
			b, err := ioutil.ReadFile(filepath.Join(directoryPath, file.Name()))
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

	resp, _, err := provider.sendv1Query(seedQuery)
	if err != nil {
		return errors.Wrap(err, "error running V1 hasura query")
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New("error executing V1 hasura query")
	}

	return nil
}
