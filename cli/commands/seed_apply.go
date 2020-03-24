package commands

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/hasura/graphql-engine/cli"
	"github.com/hasura/graphql-engine/cli/seed"

	v1 "github.com/hasura/graphql-engine/cli/client/v1"
)

type seedApplyOptions struct {
	ec *cli.ExecutionContext

	// seed file to apply
	fileName string
}

func newSeedApplyCmd(ec *cli.ExecutionContext) *cobra.Command {
	opts := seedApplyOptions{
		ec: ec,
	}
	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply seeds on the datase",
		Example: `  # Apply all seeds on the database:
  hasura seed apply

  # Apply only a particular file:
  hasura seed apply --file seeds/1234_new_table.sql`,
		SilenceUsage: false,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return ec.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.ec.Spin("Applying seeds...")
			err := opts.run()
			opts.ec.Spinner.Stop()
			if err != nil {
				return err
			}
			opts.ec.Logger.Info("Seeds planted")
			return nil
		},
	}
	cmd.Flags().StringVarP(&opts.fileName, "file", "f", "", "seed file to apply")

	return cmd
}

func (o *seedApplyOptions) run() error {
	client, err := v1.NewClient(o.ec.Config.Endpoint, nil, map[string]string{
		XHasuraAdminSecret: o.ec.Config.AdminSecret,
	})
	if err != nil {
		return err
	}
	fs := afero.NewOsFs()
	return seed.ApplySeedsToDatabase(fs, client, o.ec.SeedsDirectory, o.fileName)
}
