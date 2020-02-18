package commands

import (
	"fmt"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/hasura/graphql-engine/cli"
	"github.com/hasura/graphql-engine/cli/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			if ec.Config.Version != cli.V2 {
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

func (o *actionsUseCodegenOptions) run() error {
	o.EC.Spin("Ensuring codegen-assets repo is updated...")
	defer o.EC.Spinner.Stop()
	// ensure the the actions-codegen repo is updated
	err := o.EC.CodegenAssetsRepo.EnsureUpdated()
	if err != nil {
		o.EC.Logger.Warnf("unable to update codegen-assets repo, got %v", err)
	}

	newCodegenExecutionConfig := o.EC.Config.ActionConfig.Codegen
	newCodegenExecutionConfig.Framework = ""

	o.EC.Spin("Fetching frameworks...")
	allFrameworks, err := getCodegenFrameworks()
	if err != nil {
		return errors.Wrap(err, "error in fetching codegen frameworks")
	}

	if o.framework == "" {
		// if framework flag is not provided, display a list and allow them to choose
		var frameworkList []string
		for _, f := range allFrameworks {
			frameworkList = append(frameworkList, f.Name)
		}
		sort.Strings(frameworkList)
		o.EC.Spinner.Stop()
		newCodegenExecutionConfig.Framework, err = util.GetSelectPrompt("Choose a codegen framework to use", frameworkList)
		if err != nil {
			return errors.Wrap(err, "error in selecting framework")
		}
	} else {
		for _, f := range allFrameworks {
			if o.framework == f.Name {
				newCodegenExecutionConfig.Framework = o.framework
			}
		}
		if newCodegenExecutionConfig.Framework == "" {
			return fmt.Errorf("framework %s is not found", o.framework)
		}
	}

	hasStarterKit := false
	for _, f := range allFrameworks {
		if f.Name == newCodegenExecutionConfig.Framework && f.HasStarterKit {
			hasStarterKit = true
		}
	}

	// if with-starter-kit flag is set and the same is not available for the framework, return error
	if o.withStarterKit && !hasStarterKit {
		return fmt.Errorf("starter kit is not available for framework %s", newCodegenExecutionConfig.Framework)
	}

	// if with-starter-kit flag is not provided, give an option to clone a starterkit
	if !o.withStarterKit && hasStarterKit {
		shouldCloneStarterKit, err := util.GetYesNoPrompt("Do you also want to clone a starter kit for " + newCodegenExecutionConfig.Framework + "?")
		if err != nil {
			return err
		}
		o.withStarterKit = shouldCloneStarterKit == "y"
	}

	// if output directory is not provided, make them enter it
	if o.outputDir == "" {
		outputDir, err := util.GetFSPathPrompt("Where do you want to place the codegen files?", o.EC.Config.ActionConfig.Codegen.OutputDir)
		if err != nil {
			return errors.Wrap(err, "error in getting output directory input")
		}
		newCodegenExecutionConfig.OutputDir = outputDir
	} else {
		newCodegenExecutionConfig.OutputDir = o.outputDir
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
			filepath.Join(o.EC.GlobalConfigDir, util.ActionsCodegenDirName, newCodegenExecutionConfig.Framework, "starter-kit"),
			destinationDir,
		)
		if err != nil {
			return errors.Wrap(err, "error in copying starter kit")
		}
		o.EC.Logger.Info("Starter kit cloned at " + destinationDir)
	}

	newConfig := o.EC.Config
	newConfig.ActionConfig.Codegen = newCodegenExecutionConfig
	err = o.EC.WriteConfig(newConfig)
	if err != nil {
		return errors.Wrap(err, "error in writing config")
	}
	o.EC.Logger.Info("Codegen configuration updated in config.yaml")
	return nil
}
