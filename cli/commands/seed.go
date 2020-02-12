package commands

import (
	"github.com/hasura/graphql-engine/cli"
	"github.com/spf13/cobra"
)

// NewSeedCmd will return the seed command
func NewSeedCmd(ec *cli.ExecutionContext) *cobra.Command {
	seedCmd := &cobra.Command{
		Use:          "seed",
		Short:        "work with seed data",
		Long:         "",
		SilenceUsage: true,
	}

	seedCmd.AddCommand(
		newSeedCreateCmd(ec),
		newSeedApplyCmd(ec),
	)

	return seedCmd
}
