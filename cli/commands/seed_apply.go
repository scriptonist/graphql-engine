package commands

import (
	"github.com/hasura/graphql-engine/cli"
	"github.com/hasura/graphql-engine/cli/seed"
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
	hasuraV1APIProvider, err := seed.NewHasuraV1APIProvider(o.ec.ServerConfig.Endpoint)
	if err != nil {
		return err
	}
	return seed.ApplySeedsToDatabase(o.ec.SeedsDirectory, hasuraV1APIProvider)
}