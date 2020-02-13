package commands

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/hasura/graphql-engine/cli"
	"github.com/hasura/graphql-engine/cli/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	actionsCodegenRepoURI string = "https://github.com/hasura/codegen-assets.git"
	actionsCodegenDirName string = "actions-codegen-assets"
)

type codegenFramework struct {
	Name          string `json:"name"`
	HasStarterKit bool   `json:"hasStarterKit"`
}

func newActionsUseCodegenCmd(ec *cli.ExecutionContext) *cobra.Command {
	v := viper.New()
	opts := &actionsUseCodegenOptions{
		EC: ec,
	}
	actionsUseCodegenCmd := &cobra.Command{
		Use:          "use-codegen",
		Short:        "",
		SilenceUsage: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			ec.Viper = v
			err := ec.Prepare()
			if err != nil {
				return err
			}
			err = ec.Validate()
			if err != nil {
				return err
			}
			if ec.Config.Version == "1" {
				return fmt.Errorf("actions commands can be executed only when config version is greater than 1")
			}
			if ec.MetadataDir == "" {
				return fmt.Errorf("actions commands can be executed only when metadata_dir is set in config")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return opts.run()
		},
	}

	f := actionsUseCodegenCmd.Flags()

	f.StringVar(&opts.framework, "framework", "", "")
	f.StringVar(&opts.outputDir, "output-dir", "", "")
	f.BoolVar(&opts.withStarterKit, "with-starter-kit", false, "")

	f.String("endpoint", "", "http(s) endpoint for Hasura GraphQL Engine")
	f.String("admin-secret", "", "admin secret for Hasura GraphQL Engine")
	f.String("access-key", "", "access key for Hasura GraphQL Engine")
	f.MarkDeprecated("access-key", "use --admin-secret instead")

	// need to create a new viper because https://github.com/spf13/viper/issues/233
	v.BindPFlag("endpoint", f.Lookup("endpoint"))
	v.BindPFlag("admin_secret", f.Lookup("admin-secret"))
	v.BindPFlag("access_key", f.Lookup("access-key"))

	return actionsUseCodegenCmd
}

type actionsUseCodegenOptions struct {
	EC *cli.ExecutionContext

	framework      string
	outputDir      string
	withStarterKit bool
}

func (o *actionsUseCodegenOptions) run() (err error) {
	// ensure the the actions-codegen repo is cloned and updated

	o.framework = strings.TrimSpace(o.framework)
	o.outputDir = strings.TrimSpace(o.outputDir)

	o.EC.Spin("Fetching codegen assets...")
	defer o.EC.Spinner.Stop()
	actionsCodegenGit := util.NewGitUtil(
		actionsCodegenRepoURI,
		filepath.Join(o.EC.GlobalConfigDir, actionsCodegenDirName),
		"",
	)
	err = actionsCodegenGit.EnsureCloned()
	if err != nil {
		return
	}
	_ = actionsCodegenGit.EnsureUpdated()
	newCodegenExecutionConfig := o.EC.Config.Action.Codegen
	o.EC.Spinner.Stop()

	frameworkFileBytes, err := ioutil.ReadFile(filepath.Join(o.EC.GlobalConfigDir, actionsCodegenDirName, "frameworks.json"))
	if err != nil {
		return err
	}
	var allFrameworks []codegenFramework
	err = json.Unmarshal(frameworkFileBytes, &allFrameworks)
	if err != nil {
		return err
	}

	// if framework flag is not provided, display a list and allow them to choose
	if o.framework == "" {
		var frameworkList []string
		for _, f := range allFrameworks {
			frameworkList = append(frameworkList, f.Name)
		}
		sort.Strings(frameworkList)
		newCodegenExecutionConfig.Framework, err = util.GetSelectPrompt("Choose a codegen framework to use", frameworkList)
		if err != nil {
			return err
		}
	} else {
		newCodegenExecutionConfig.Framework = o.framework
	}

	hasStarterKit := false
	for _, f := range allFrameworks {
		if f.Name == newCodegenExecutionConfig.Framework && f.HasStarterKit {
			hasStarterKit = true
		}
	}

	// if with-starter-kit flag is not provided, give an option to clone a starterkit
	if !o.withStarterKit && hasStarterKit {
		shouldCloneStarterKit, err := util.GetYesNoPrompt("Do you also want to clone a starter kit for " + newCodegenExecutionConfig.Framework + "?")
		if err != nil {
			return err
		}
		o.withStarterKit = shouldCloneStarterKit == "y"
	}

	// clone the starter kit
	if o.withStarterKit && hasStarterKit {
		// get a directory name to clone the starter kit in
		starterKitDirname := newCodegenExecutionConfig.Framework
		err = util.FSCheckIfDirPathExists(
			filepath.Join(o.EC.ExecutionDirectory, starterKitDirname),
		)
		suffix := 2
		for err == nil {
			starterKitDirname = newCodegenExecutionConfig.Framework + "-" + strconv.Itoa(suffix)
			suffix++
			err = util.FSCheckIfDirPathExists(starterKitDirname)
		}
		err = nil

		// copy the starter kit
		destinationDir := filepath.Join(o.EC.ExecutionDirectory, starterKitDirname)
		err = util.FSCopyDir(
			filepath.Join(o.EC.GlobalConfigDir, actionsCodegenDirName, newCodegenExecutionConfig.Framework, "starter-kit"),
			destinationDir,
		)

		if err != nil {
			return err
		}

		o.EC.Logger.Info("Starter kit cloned at " + destinationDir)
	}

	// if output directory is not provided, make them enter it
	if o.outputDir == "" {
		outputDir, err := util.GetFSPathPrompt("Where do you want to place the codegen files?", o.EC.Config.Action.Codegen.OutputDir)
		if err != nil {
			return err
		}
		newCodegenExecutionConfig.OutputDir = outputDir
	} else {
		newCodegenExecutionConfig.OutputDir = o.outputDir
	}

	newConfig := o.EC.Config
	newConfig.Action.Codegen = newCodegenExecutionConfig

	configString, err := yaml.Marshal(newConfig)
	if err != nil {
		return
	}

	err = ioutil.WriteFile(o.EC.ConfigFile, configString, 0644)
	if err != nil {
		return
	}

	o.EC.Logger.Info("Codegen configuration updated in config.yaml\n\n")

	return nil
}