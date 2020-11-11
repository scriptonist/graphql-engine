package commands

import (
	"fmt"

	"github.com/hasura/graphql-engine/cli/util"

	"github.com/hasura/graphql-engine/cli/migrate"

	"github.com/hasura/graphql-engine/cli/internal/scripts"
	"github.com/spf13/afero"

	"github.com/hasura/graphql-engine/cli"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newUpdateMultipleSources(ec *cli.ExecutionContext) *cobra.Command {
	v := viper.New()
	cmd := &cobra.Command{
		Use:   "update-project-multiple-sources",
		Short: "update project to use multiple datasources",
		Long: `
`,
		Example:      `  `,
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ec.Viper = v
			err := ec.Prepare()
			if err != nil {
				return err
			}
			return ec.Validate()
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			// preconditions
			// project should be using Config V2
			// server should be >= v1.4
			case ec.Config.Version != cli.V2:
				return fmt.Errorf("project should be using config V2 to be able to use multiple datasources")
			case !ec.Version.ServerFeatureFlags.HasDatasources:
				return fmt.Errorf("server doesn't support multiple data sources")
			}
			//get the list of data sources
			migrateDrv, err := migrate.NewMigrate(ec, true)
			if err != nil {
				return err
			}
			datasources, err := migrateDrv.GetDatasources()
			if err != nil {
				return err
			}
			targetDatasource, err := util.GetSelectPrompt("select datasource for which current migrations belong to", datasources)
			if err != nil {
				return err
			}
			fmt.Println(datasources)
			opts := scripts.UpgradeToMuUpgradeProjectToMultipleSourcesOpts{
				Fs:                   afero.NewOsFs(),
				ProjectDirectory:     ec.ExecutionDirectory,
				MigrationsDirectory:  ec.MigrationDir,
				TargetDatasourceName: targetDatasource,
				Logger:               ec.Logger,
			}
			if err := scripts.UpgradeProjectToMultipleSources(opts); err != nil {
				return err
			}
			return nil
		},
	}
	return cmd
}