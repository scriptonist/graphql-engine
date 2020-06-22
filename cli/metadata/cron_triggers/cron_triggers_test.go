package crontriggers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hasura/graphql-engine/cli/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

func TestCronTriggers_Build(t *testing.T) {
	type fields struct {
		fs                 afero.Fs
		MetadataDir        string
		logger             *logrus.Logger
		serverFeatureFlags *version.ServerFeatureFlags
	}
	type args struct {
		metadata *yaml.MapSlice
	}
	fileContents := [][]byte{
		[]byte(`
- name: test
  webhook: https://httpbin.org
  schedule: '*/10 * * * *'
  include_in_metadata: true
  payload:
    test: test
`),
		[]byte(`[]`),
		[]byte(`
- name: test
  webhook: https://httpbin.org
  schedule: '*/10 * * * *'
  include_in_metadata: true
  payload:
    test: test
- name: test2
  webhook: https://httpbin.org
  schedule: '*/10 */10 * * *'
  include_in_metadata: true
  payload:
    test: test
`),
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"can build cron triggers when a file is present in metadata dir",
			fields{
				func() afero.Fs {
					fs := afero.NewMemMapFs()
					contents := fileContents[0]
					err := afero.WriteFile(fs, "metadata/cron_triggers.yaml", contents, 0655)
					if err != nil {
						t.Fatal(err)
					}
					return fs
				}(),
				"metadata",
				logrus.New(),
				&version.ServerFeatureFlags{
					HasCronTriggers: true,
				},
			},
			args{
				func() *yaml.MapSlice {
					var y = new(yaml.MapSlice)
					return y
				}(),
			},
			false,
		},
		{
			"assume an empty file when cron_triggers.yaml file is not present",
			fields{
				func() afero.Fs {
					fs := afero.NewMemMapFs()
					assert.NoError(t, afero.WriteFile(fs, "metadata/test.yaml", []byte("nothing"), 0755))
					return fs
				}(),
				"metadata",
				logrus.New(),
				&version.ServerFeatureFlags{
					HasCronTriggers: true,
				},
			},
			args{
				func() *yaml.MapSlice {
					var y yaml.MapSlice
					return &y
				}(),
			},
			false,
		},
		{
			"can build cron triggers correctly when multiple cron triggers are present in file",
			fields{
				func() afero.Fs {
					fs := afero.NewMemMapFs()
					contents := fileContents[2]
					err := afero.WriteFile(fs, "metadata/cron_triggers.yaml", contents, 0655)
					if err != nil {
						t.Fatal(err)
					}
					return fs
				}(),
				"metadata",
				logrus.New(),
				&version.ServerFeatureFlags{
					HasCronTriggers: true,
				},
			},
			args{
				func() *yaml.MapSlice {
					var y = new(yaml.MapSlice)
					return y
				}(),
			},
			false,
		},
	}
	for testind, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CronTriggers{
				FS:                 tt.fields.fs,
				MetadataDir:        tt.fields.MetadataDir,
				logger:             tt.fields.logger,
				serverFeatureFlags: tt.fields.serverFeatureFlags,
			}
			err := c.Build(tt.args.metadata)
			if !tt.wantErr {
				assert.NoError(t, err)
			}
			var wantMetadata = yaml.MapSlice{}
			var want = yaml.MapItem{
				Key:   metadataKey,
				Value: yaml.MapSlice{},
			}
			err = yaml.Unmarshal(fileContents[testind], &want.Value)
			assert.NoError(t, err)
			assert.Equal(t, append(wantMetadata, want), *tt.args.metadata)
		})
	}
}
