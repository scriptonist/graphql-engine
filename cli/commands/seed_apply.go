package commands

import (
	"github.com/hasura/graphql-engine/cli"
	v1 "github.com/hasura/graphql-engine/cli/client/v1"
	"github.com/hasura/graphql-engine/cli/seed"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

type seedApplyOptions struct {
	ec *cli.ExecutionContext
}

func newSeedApplyCmd(ec *cli.ExecutionContext) *cobra.Command {
	opts := seedApplyOptions{
		ec: ec,
	}
	cmd := &cobra.Command{
		Use:          "apply",
		Short:        "apply seed",
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
	return cmd
}

func (o *seedApplyOptions) run() error {
	client, err := v1.NewClient(o.ec.Config.Endpoint, map[string]string{
		XHasuraAdminSecret: o.ec.Config.AdminSecret,
	})
	if err != nil {
		return err
	}
	fs := afero.NewOsFs()
	return seed.ApplySeedsToDatabase(fs, client, o.ec.SeedsDirectory)
}
