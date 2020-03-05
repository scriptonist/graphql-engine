package commands

import (
	"github.com/docker/docker/pkg/namesgenerator"
	"github.com/hasura/graphql-engine/cli"
	v1 "github.com/hasura/graphql-engine/cli/client/v1"
	"github.com/hasura/graphql-engine/cli/metadata/actions/editor"
	"github.com/hasura/graphql-engine/cli/migrate/database/hasuradb"
	"github.com/hasura/graphql-engine/cli/seed"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type seedNewOptions struct {
	ec *cli.ExecutionContext

	// filename for the new seed file
	seedname string
	// table name if seed file has to be created from a database table
	fromTableNames []string
}

func newSeedCreateCmd(ec *cli.ExecutionContext) *cobra.Command {
	opts := seedNewOptions{
		ec: ec,
	}
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "create a new seed file",
		SilenceUsage: false,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ec.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.seedname = args[0]
			} else {
				opts.seedname = namesgenerator.GetRandomName(0)
			}
			err := opts.run()
			if err != nil {
				return err
			}
			ec.Logger.Info("created seed file successfully")
			return nil

		},
	}

	cmd.Flags().StringArrayVar(&opts.fromTableNames, "from-tables", []string{}, "name of table from which seed file has to be initialized")

	return cmd
}

func (o *seedNewOptions) run() error {
	createSeedOpts := seed.CreateSeedOpts{
		UserProvidedSeedName: o.seedname,
		DirectoryPath:        o.ec.SeedsDirectory,
	}

	// If we are initializing from a database table
	// create a hasura client and add table name opts
	if len(o.fromTableNames) > 0 {
		// Create a client
		client, err := v1.NewClient(o.ec.Config.Endpoint, map[string]string{
			XHasuraAdminSecret: o.ec.Config.AdminSecret,
		})
		if err != nil {
			return errors.Wrap(err, "cannot initialize hasura client")
		}
		// Send a pg dump query to dump just sql
		pgDumpOpts := []string{"--no-owner", "--no-acl", "--data-only", "--inserts"}
		for _, table := range o.fromTableNames {
			pgDumpOpts = append(pgDumpOpts, "--table", table)
		}
		query := hasuradb.SchemaDump{
			Opts:        pgDumpOpts,
			CleanOutput: true,
		}
		// Send the query
		_, body, err := client.SendPGDumpQuery(query)
		if err != (*v1.Error)(nil) {
			return errors.Wrap(err, "error executing operation")
		}
		createSeedOpts.Data = body
	} else {
		const defaultText = ""
		data, err := editor.CaptureInputFromEditor(editor.GetPreferredEditorFromEnvironment, defaultText, "*.sql")
		if err != nil {
			return err
		}
		createSeedOpts.Data = data
	}

	fs := afero.NewOsFs()
	filepath, err := seed.CreateSeedFile(fs, createSeedOpts)
	if err != nil || filepath == nil {
		return errors.Wrap(err, "failed to create seed file")
	}

	return nil
}
