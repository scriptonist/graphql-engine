package seed

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"testing"

	v1 "github.com/hasura/graphql-engine/cli/client/v1"
	"github.com/spf13/afero"
)

func TestCreateSeedFile(t *testing.T) {
	if !*hasura {
		t.Skip()
	}
	client, err := v1.NewClient("http://localhost:8080")
	if err != nil {
		t.Fatalf("cannot create client: %v", err)
	}
	// Add test data
	createTestDataQuery := map[string]interface{}{
		"type": "bulk",
		"args": []map[string]interface{}{
			// Create table 1
			{
				"type": "run_sql",
				"args": map[string]string{
					"sql": `
					CREATE TABLE account(
						user_id serial PRIMARY KEY,
						username VARCHAR (50) UNIQUE NOT NULL,
						password VARCHAR (50) NOT NULL,
						email VARCHAR (355) UNIQUE NOT NULL);`,
				},
			},
			// Insert data
			{
				"type": "run_sql",
				"args": map[string]string{
					"sql": `INSERT INTO account (username, password, email) values ('scriptonist', 'no you cant guess it', 'hello@drogon.com');`,
				},
			},
			// Create table 2
			{
				"type": "run_sql",
				"args": map[string]string{
					"sql": `
					CREATE TABLE account2(
						user_id serial PRIMARY KEY,
						username VARCHAR (50) UNIQUE NOT NULL,
						password VARCHAR (50) NOT NULL,
						email VARCHAR (355) UNIQUE NOT NULL);`,
				},
			},
			// Insert data
			{
				"type": "run_sql",
				"args": map[string]string{
					"sql": `INSERT INTO account2 (username, password, email) values ('scriptonist', 'no you cant guess it', 'hello@drogon.com');`,
				},
			},
			// Create table 3
			{
				"type": "run_sql",
				"args": map[string]string{
					"sql": `
					CREATE TABLE account3(
						user_id serial PRIMARY KEY,
						username VARCHAR (50) UNIQUE NOT NULL,
						password VARCHAR (50) NOT NULL,
						email VARCHAR (355) UNIQUE NOT NULL);`,
				},
			},
			// Insert data
			{
				"type": "run_sql",
				"args": map[string]string{
					"sql": `INSERT INTO account3 (username, password, email) values ('scriptonist', 'no you cant guess it', 'hello@drogon.com');`,
				},
			},
		},
	}

	r, rbody, err := client.SendQuery(createTestDataQuery)
	if err != nil || r.StatusCode != http.StatusOK {
		t.Fatalf("Cannot initialize testdata: %v %v", err, string(rbody))
	}

	type args struct {
		fs   afero.Fs
		opts CreateSeedOpts
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
		// used when creating from tables
		wantTableSQL string
	}{
		{
			name: "can create file",
			args: args{
				fs: afero.NewMemMapFs(),
				opts: CreateSeedOpts{
					DirectoryPath:        "seeds/",
					UserProvidedSeedName: "xyz",
				},
			},
			wantErr: false,
		},
		{
			name: "can we create seed file from table",
			args: args{
				fs: afero.NewMemMapFs(),
				opts: CreateSeedOpts{
					DirectoryPath:        "seeds/",
					UserProvidedSeedName: "xyzfromtable",
					CreateFromTableOpts: &CreateFromTableOpts{
						TableNames:   []string{"account"},
						PGDumpClient: client.ClientPGDump,
					},
				},
			},
			wantTableSQL: `INSERT INTO public.account VALUES (1, 'scriptonist', 'no you cant guess it', 'hello@drogon.com');
SELECT pg_catalog.setval('public.account_user_id_seq', 1, true);
`,
			wantErr: false,
		},
		{
			name: "when creating seed from multiple tables are the seeds merged to a single file",
			args: args{
				fs: afero.NewMemMapFs(),
				opts: CreateSeedOpts{
					DirectoryPath:        "seeds/",
					UserProvidedSeedName: "getfromthreetables",
					CreateFromTableOpts: &CreateFromTableOpts{
						TableNames:   []string{"account", "account2", "account3"},
						PGDumpClient: client.ClientPGDump,
					},
				},
			},
			wantTableSQL: `INSERT INTO public.account VALUES (1, 'scriptonist', 'no you cant guess it', 'hello@drogon.com');
INSERT INTO public.account2 VALUES (1, 'scriptonist', 'no you cant guess it', 'hello@drogon.com');
INSERT INTO public.account3 VALUES (1, 'scriptonist', 'no you cant guess it', 'hello@drogon.com');
SELECT pg_catalog.setval('public.account2_user_id_seq', 1, true);
SELECT pg_catalog.setval('public.account3_user_id_seq', 1, true);
SELECT pg_catalog.setval('public.account_user_id_seq', 1, true);
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateSeedFile(tt.args.fs, tt.args.opts)
			if err != nil && tt.wantErr == false {
				t.Fatalf("CreateSeedFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Fatalf("Creating seed file failed %v", err)
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
			// See if creating seed from tables succeded
			if tt.args.opts.CreateFromTableOpts != nil {
				var fileFound = false
				afero.Walk(tt.args.fs, tt.args.opts.DirectoryPath, func(path string, info os.FileInfo, err error) error {
					var re = regexp.MustCompile(tt.args.opts.UserProvidedSeedName)
					if !info.IsDir() && re.Match([]byte(info.Name())) {
						fileFound = true
						b, err := afero.ReadFile(tt.args.fs, path)
						if err != nil {
							t.Errorf("error while reading seed file: %v", err)
						}
						if string(b) != tt.wantTableSQL {
							fmt.Println("-----")
							fmt.Println(string(b))
							fmt.Println("-----")
							fmt.Println()
							fmt.Println("-----")
							fmt.Println(tt.wantTableSQL)
							fmt.Println("-----")
							fmt.Println()
							t.Fatalf("Filename: %v Want: %v, got: %v", path, tt.wantTableSQL, string(b))
						}
					}
					return err

				})
				if !fileFound {
					t.Fatalf("seed file not created for %v", tt.name)
				}
			}
		})
	}
}
