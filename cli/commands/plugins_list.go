package commands

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/hasura/graphql-engine/cli"
	"github.com/hasura/graphql-engine/cli/plugins/index"
	"github.com/hasura/graphql-engine/cli/plugins/installation"
	"github.com/hasura/graphql-engine/cli/plugins/types"
)

func newPluginsListCmd(ec *cli.ExecutionContext) *cobra.Command {
	pluginsListCmd := &cobra.Command{
		Use:          "list",
		Short:        "",
		Example:      ``,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ec.Spin("Fetching plugins list...")
			defer ec.Spinner.Stop()
			plugins, err := index.LoadPluginListFromFS(ec.PluginsPath.IndexPluginsPath())
			if err != nil {
				return errors.Wrap(err, "failed to load the list of plugins from the index")
			}
			names := make([]string, len(plugins))
			pluginMap := make(map[string]types.Plugin, len(plugins))
			for i, p := range plugins {
				names[i] = p.Name
				pluginMap[p.Name] = p
			}

			installed, err := installation.ListInstalledPlugins(ec.PluginsPath.InstallReceiptsPath())
			if err != nil {
				return errors.Wrap(err, "failed to load installed plugins")
			}

			// No plugins found
			if len(names) == 0 {
				return nil
			}

			var rows [][]string
			cols := []string{"NAME", "DESCRIPTION", "INSTALLED"}
			for _, name := range names {
				plugin := pluginMap[name]
				var status string
				if _, ok := installed[name]; ok {
					status = "yes"
				} else if _, ok, err := installation.GetMatchingPlatform(plugin.Platforms); err != nil {
					return errors.Wrapf(err, "failed to get the matching platform for plugin %s", name)
				} else if ok {
					status = "no"
				} else {
					status = "unavailable on " + runtime.GOOS
				}
				rows = append(rows, []string{name, limitString(plugin.ShortDescription, 50), status})
			}
			rows = sortByFirstColumn(rows)
			ec.Spinner.Stop()
			return printTable(os.Stdout, cols, rows)
		},
	}
	return pluginsListCmd
}

func printTable(out io.Writer, columns []string, rows [][]string) error {
	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprint(w, strings.Join(columns, "\t"))
	fmt.Fprintln(w)
	for _, values := range rows {
		fmt.Fprint(w, strings.Join(values, "\t"))
		fmt.Fprintln(w)
	}
	return w.Flush()
}

func sortByFirstColumn(rows [][]string) [][]string {
	sort.Slice(rows, func(a, b int) bool {
		return rows[a][0] < rows[b][0]
	})
	return rows
}