package seed

import (
	"regexp"
	"testing"

	"github.com/spf13/afero"
)

func TestCreateSeedFile1(t *testing.T) {
	type args struct {
		fs   afero.Fs
		opts CreateSeedOpts
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "can we create seed files",
			args: args{fs: afero.NewMemMapFs(),
				opts: CreateSeedOpts{
					DirectoryPath:        "seeds/",
					Data:                 []byte("INSERT INTO account1 (username, password, email) values ('scriptonist', 'no you cant guess it', 'hello@drogon.com');"),
					UserProvidedSeedName: "can_we_create_seed_files",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateSeedFile(tt.args.fs, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateSeedFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Do a regex match for filename returned
			// check if it is in required format
			var re = regexp.MustCompile(`^([a-z]+\/)([0-9]+)\_(.+)(\.sql)$`)
			regexGroups := re.FindStringSubmatch(*got)

			// Since filename has to be in form
			// dirname/21212_filename.sql
			// regexGroups should have 5 elements
			// element 0: whole string
			// element 1: dirname
			// element 2: timestamp
			// element 3: filename
			// element 4: extension
			if len(regexGroups) != 5 {
				t.Fatalf("CreateSeedFile() = %v, but want filepath of form"+` [a-z]+\/[0-9]+\_[a-zA-Z]+\.sql`, *got)
			}
			gotDirectoryPath := regexGroups[1]
			gotUserProvidedFilename := regexGroups[3]
			gotFileExtension := regexGroups[4]
			if gotDirectoryPath != tt.args.opts.DirectoryPath {
				t.Errorf("CreateSeedFile() = %v, but want directory path %s , got %s ", *got, tt.args.opts.DirectoryPath, gotDirectoryPath)
			}
			if gotUserProvidedFilename != tt.args.opts.UserProvidedSeedName {
				t.Errorf("CreateSeedFile() = %v, but want filename %s , got %s ", *got, tt.args.opts.UserProvidedSeedName, gotUserProvidedFilename)
			}
			if gotFileExtension != ".sql" {
				t.Errorf("CreateSeedFile() = %v, want fileextension .sql got %s", *got, gotFileExtension)
			}

			// test if a filewith the filename was created
			if s, err := tt.args.fs.Stat(*got); err != nil {
				if s.IsDir() {
					t.Fatalf("expected to get a file with name %v", *got)
				}
			}
		})
	}
}
