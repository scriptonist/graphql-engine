package seed

import (
	"flag"
	"testing"

	v1 "github.com/hasura/graphql-engine/cli/client/v1"

	"github.com/spf13/afero"
)

func TestApplySeedsToDatabase(t *testing.T) {
	flag.Parse()
	if !*hasura {
		// This test want a running hasura instance
		t.Skip()
	}
	client, err := v1.NewClient("http://localhost:8080")
	if err != nil {
		t.Fatalf("Cannot create hasura client: %v", err)
	}

	type args struct {
		fs                afero.Fs
		hasuraAPIProvider HasuraAPIProvider
		directoryPath     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "can apply a seed",
			args: args{
				directoryPath: "seeds/",
				fs: func(fs afero.Fs) afero.Fs {
					var sql = `
						CREATE TABLE account(
							id serial PRIMARY KEY,
							username VARCHAR (50) UNIQUE NOT NULL
						);
						
						INSERT INTO account (username) values ('test_user');
						`
					err = afero.WriteFile(fs, "seeds/seed.sql", []byte(sql), 0655)
					if err != nil {
						t.Fatalf("cannot create seed files: %v", err)
					}
					return fs
				}(afero.NewMemMapFs()),
				hasuraAPIProvider: client,
			},
			wantErr: false,
		},
		{
			name: "can throw error when bad SQL is given",
			args: args{
				directoryPath: "seeds/",
				fs: func(fs afero.Fs) afero.Fs {
					afero.WriteFile(fs, "seeds/bad.sql", []byte("insert into gibberish gibberish"), 0655)
					return fs
				}(afero.NewMemMapFs()),
				hasuraAPIProvider: client,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ApplySeedsToDatabase(tt.args.fs, tt.args.hasuraAPIProvider, tt.args.directoryPath); (err != nil) != tt.wantErr {
				t.Errorf("ApplySeedsToDatabase() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
