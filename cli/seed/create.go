package seed

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hasura/graphql-engine/cli/migrate/database/hasuradb"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

type CreateFromTableOpts struct {
	TableNames   []string
	PGDumpClient interface {
		SendPGDumpQuery(m interface{}) (*http.Response, []byte, error)
	}
}

// CreateSeedOpts has the list of options required
// to create a seed file
type CreateSeedOpts struct {
	UserProvidedSeedName string
	// DirectoryPath in which seed file should be created
	DirectoryPath       string
	CreateFromTableOpts *CreateFromTableOpts
}

// CreateSeedFile creates a .sql file according to the arguments
// it'll return full filepath and an error if any
func CreateSeedFile(fs afero.Fs, opts CreateSeedOpts) (*string, error) {
	const fileExtension = "sql"

	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	// filename will be in format <timestamp>_<userProvidedSeedName>.sql
	filenameWithTimeStamp := fmt.Sprintf("%s_%s.%s", timestamp, opts.UserProvidedSeedName, fileExtension)
	fullFilePath := filepath.Join(opts.DirectoryPath, filenameWithTimeStamp)

	// Create file
	file, err := fs.Create(fullFilePath)
	if err != nil {
		return nil, err
	}

	// See if contents has to be populated from table
	if opts.CreateFromTableOpts != nil {
		if opts.CreateFromTableOpts.PGDumpClient == nil {
			return nil, errors.Errorf("pgdump client not provided")
		}
		// Send a pg dump query to dump just sql
		pgDumpOpts := []string{"--no-owner", "--no-acl", "--data-only", "--inserts"}
		for _, table := range opts.CreateFromTableOpts.TableNames {
			pgDumpOpts = append(pgDumpOpts, "--table", table)
		}
		query := hasuradb.SchemaDump{
			Opts:        pgDumpOpts,
			CleanOutput: true,
		}
		// Send the query
		resp, body, err := opts.CreateFromTableOpts.PGDumpClient.SendPGDumpQuery(query)

		var horror hasuradb.HasuraError
		if resp.StatusCode != http.StatusOK {
			err = json.Unmarshal(body, &horror)
			if err != nil {
				return nil, err
			}
			return nil, errors.New(string(body))
		}
		if err != nil {
			return nil, errors.New(string(body))
		}

		// If not error write the body to file
		err = afero.WriteFile(fs, fullFilePath, body, 0655)
		if err != nil {
			return nil, err
		}

	}
	defer file.Close()
	return &fullFilePath, nil
}
