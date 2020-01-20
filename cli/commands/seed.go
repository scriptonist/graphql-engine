package commands

import (
	"sync"

	"github.com/hasura/graphql-engine/cli"
	"github.com/spf13/cobra"
)

// NewSeedCmd will return the seed command
func NewSeedCmd(ec *cli.ExecutionContext) *cobra.Command {
	// v := viper.New()

	/*
		opts := &seedCmdOptions{
			EC: ec,
		}
	*/
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

type seedCmdOptions struct {
	EC *cli.ExecutionContext

	APIPort     string
	ConsolePort string
	Address     string

	DontOpenBrowser bool

	WG *sync.WaitGroup

	StaticDir string
	Browser   string
}

func (o *seedCmdOptions) run() error {
	// log := o.EC.Logger
	return nil
}
