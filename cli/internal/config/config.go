package config

import (
	"io/ioutil"

	"github.com/spf13/cobra"

	"github.com/hasura/graphql-engine/cli/util"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/hasura/graphql-engine/cli/metadata/actions/types"
)

// Config represents configuration required for the CLI to function
type Config struct {
	// Version of the config.
	Version Version `yaml:"version,omitempty"`

	// ServerConfig to be used by CLI to contact server.
	ServerConfig `yaml:",inline"`

	// MetadataDirectory defines the directory where the metadata files were stored.
	MetadataDirectory string `yaml:"metadata_directory,omitempty"`
	// MigrationsDirectory defines the directory where the migration files were stored.
	MigrationsDirectory string `yaml:"migrations_directory,omitempty"`
	// ActionConfig defines the config required to create or generate codegen for an action.
	ActionConfig *types.ActionExecutionConfig `yaml:"actions,omitempty"`
}

// readConfig reads the configuration from config file, flags and env vars,
// through viper.
func ReadConfig(v *viper.Viper, configFileName string) (*Config, error) {
	v.SetEnvPrefix(util.ViperEnvPrefix)
	v.SetEnvKeyReplacer(util.ViperEnvReplacer)
	v.AutomaticEnv()
	v.SetDefault("version", "2")
	v.SetDefault("endpoint", "http://localhost:8080")
	v.SetDefault("admin_secret", "")
	v.SetDefault("access_key", "")
	v.SetDefault("api_paths.query", "v1/query")
	v.SetDefault("api_paths.graphql", "v1/graphql")
	v.SetDefault("api_paths.config", "v1alpha1/config")
	v.SetDefault("api_paths.pg_dump", "v1alpha1/pg_dump")
	v.SetDefault("api_paths.version", "v1/version")
	v.SetDefault("metadata_directory", "metadata")
	v.SetDefault("migrations_directory", "migrations")
	v.SetDefault("actions.kind", "synchronous")
	v.SetDefault("actions.handler_webhook_baseurl", "http://localhost:3000")
	v.SetDefault("actions.codegen.framework", "")
	v.SetDefault("actions.codegen.output_dir", "")
	v.SetDefault("actions.codegen.uri", "")
	adminSecret := v.GetString("admin_secret")
	if adminSecret == "" {
		adminSecret = v.GetString("access_key")
	}
	config := &Config{
		Version: Version(v.GetInt("version")),
		ServerConfig: ServerConfig{
			Endpoint:    v.GetString("endpoint"),
			AdminSecret: adminSecret,
			APIPaths: &ServerAPIPaths{
				Query:   v.GetString("api_paths.query"),
				GraphQL: v.GetString("api_paths.graphql"),
				Config:  v.GetString("api_paths.config"),
				PGDump:  v.GetString("api_paths.pg_dump"),
				Version: v.GetString("api_paths.version"),
			},
			InsecureSkipTLSVerify: v.GetBool("insecure_skip_tls_verify"),
			CAPath:                v.GetString("certificate_authority"),
		},
		MetadataDirectory:   v.GetString("metadata_directory"),
		MigrationsDirectory: v.GetString("migrations_directory"),
		ActionConfig: &types.ActionExecutionConfig{
			Kind:                  v.GetString("actions.kind"),
			HandlerWebhookBaseURL: v.GetString("actions.handler_webhook_baseurl"),
			Codegen: &types.CodegenExecutionConfig{
				Framework: v.GetString("actions.codegen.framework"),
				OutputDir: v.GetString("actions.codegen.output_dir"),
				URI:       v.GetString("actions.codegen.uri"),
			},
		},
	}
	if !config.Version.IsValid() {
		return nil, ErrInvalidConfigVersion
	}
	err := config.ServerConfig.ParseEndpoint()
	if err != nil {
		return nil, errors.Wrap(err, "unable to parse server endpoint")
	}
	err = config.ServerConfig.SetTLSConfig()
	if err != nil {
		return nil, errors.Wrap(err, "setting up TLS config failed")
	}
	if err := config.ServerConfig.SetHTTPClient(); err != nil {
		return nil, errors.Wrap(err, "setting up server HTTP client")
	}
	return config, nil
}

func (c *Config) WriteConfig(newCfg *Config, to string) error {
	var cfg *Config
	if newCfg != nil {
		cfg = newCfg
	} else {
		cfg = c
	}
	y, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(to, y, 0644)
}

func LoadConfig(v *viper.Viper, cmd *cobra.Command) (*Config, error) {
	configFileName, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return nil, errors.Wrap(err, "reading config file")
	}
	c, err := ReadConfig(v, configFileName)
	if err != nil {
		return nil, err
	}

	if c.Version == V1 && cmd.Flags().Changed("config-file") {
		return nil, errors.New("config-file for config version greater than 2")
	}

	return nil, nil
}
