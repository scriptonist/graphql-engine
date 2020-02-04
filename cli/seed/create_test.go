package seed

import (
	"regexp"
	"testing"

	"github.com/spf13/afero"
)

func TestCreateSeedFile(t *testing.T) {
	type args struct {
		fs   afero.Fs
		opts CreateSeedOptions
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "can we create file",
			args: args{
				fs: afero.NewMemMapFs(),
				opts: CreateSeedOptions{
					DirectoryPath:        "seeds/",
					UserProvidedSeedName: "xyz",
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
			var re = regexp.MustCompile(`^([a-z]+\/)([0-9]+)\_([a-zA-Z]+)(\.sql)$`)
			groups := re.FindStringSubmatch(*got)
			if len(groups) != 5 {
				t.Fatalf("CreateSeedFile() = %v, but want filepath of form"+` [a-z]+\/[0-9]+\_[a-zA-Z]+\.sql`, *got)
			}
			gotDirectoryPath := groups[1]
			gotUserProvidedFilename := groups[3]
			gotFileExtension := groups[4]
			if gotDirectoryPath != tt.args.opts.DirectoryPath {
				t.Errorf("CreateSeedFile() = %v, but want directory path %s , got %s ", *got, tt.args.opts.DirectoryPath, gotDirectoryPath)
			}
			if gotUserProvidedFilename != tt.args.opts.UserProvidedSeedName {
				t.Errorf("CreateSeedFile() = %v, but want filename %s , got %s ", *got, tt.args.opts.UserProvidedSeedName, gotUserProvidedFilename)
			}
			if gotFileExtension != ".sql" {
				t.Errorf("CreateSeedFile() = %v, want fileextension .sql got %s", *got, gotFileExtension)
			}
		})
	}
}
