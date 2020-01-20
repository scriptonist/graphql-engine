package commands

import (
	"github.com/hasura/graphql-engine/cli"
	"github.com/spf13/cobra"
)

type seedCreateOptions struct {
	ec *cli.ExecutionContext
}

func newSeedCreateCmd(ec *cli.ExecutionContext) *cobra.Command {
	/*
		opts := seedCreateOptions{
			ec: ec,
		}
	*/
	cmd := &cobra.Command{
		Use:          "create",
		Short:        "create a seed file",
		SilenceUsage: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
