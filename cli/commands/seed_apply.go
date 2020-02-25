package commands

import (
	"github.com/hasura/graphql-engine/cli"
	v1 "github.com/hasura/graphql-engine/cli/client/v1"
	"github.com/hasura/graphql-engine/cli/seed"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type seedApplyOptions struct {
	ec *cli.ExecutionContext
}

func newSeedApplyCmd(ec *cli.ExecutionContext) *cobra.Command {
	opts := seedApplyOptions{
		ec: ec,
	}
	v := viper.New()
	cmd := &cobra.Command{
		Use:          "apply",
		Short:        "apply seed",
		SilenceUsage: false,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ec.Viper = v
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
	hasuraV1APIProvider, err := v1.NewClient(o.ec.Config.Endpoint)
	if err != nil {
		return err
	}
	fs := afero.NewOsFs()
	return seed.ApplySeedsToDatabase(fs, hasuraV1APIProvider.ClientMetadataAndSchema, o.ec.SeedsDirectory)
}
