package commands

import (
	"github.com/hasura/graphql-engine/cli"
	"github.com/spf13/cobra"
)

type seedApplyOptions struct {
	ec *cli.ExecutionContext
}

func newSeedApplyCmd(ec *cli.ExecutionContext) *cobra.Command {
	/*
		opts := seedCreateOptions{
			ec: ec,
		}
	*/
	cmd := &cobra.Command{
		Use:          "apply",
		Short:        "apply seed",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
