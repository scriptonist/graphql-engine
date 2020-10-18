package common

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hasura/graphql-engine/cli/internal/client"
	"github.com/hasura/graphql-engine/cli/internal/client/v1/metadata"
)

func TestCatalogState_GetCatalogState(t *testing.T) {
	type fields struct {
		client   *client.Client
		endpoint string
	}

	m := http.NewServeMux()
	ts := httptest.NewServer(m)

	tests := []struct {
		name    string
		fields  fields
		want    *metadata.GetCatalogStateResponseBody
		wantErr bool
		server  *httptest.Server
	}{
		{
			name: "can fetch and decode catalog state",
			fields: fields{
				client: func() *client.Client {
					m.Handle("/success", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var resp = []byte(`
{
    "id": "77d0a2cf-ea89-454a-9159-d24de958271b",
    "cli_state": {
        "key": "value"
    },
	"console_state": {}
}
`)

						w.Header().Set("Content-Type", "application/json")
						w.Write(resp)
					}))
					hasuraClient, err := client.NewClient(nil, fmt.Sprintf("%s/", ts.URL))
					if err != nil {
						t.Fatal(err)
					}
					return hasuraClient
				}(),
				endpoint: "success",
			},
			want: &metadata.GetCatalogStateResponseBody{
				ID: "77d0a2cf-ea89-454a-9159-d24de958271b",
				CLIState: &map[string]interface{}{
					"key": "value",
				},
				ConsoleState: &map[string]interface{}{},
			},
			wantErr: false,
		},
		{
			name: "can handle errors gracefully",
			fields: fields{
				client: func() *client.Client {
					m.Handle("/fail", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var resp = []byte(`
{
    "path": "$",
    "error": "key \"args\" not found",
    "code": "parse-failed"
}
`)

						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusBadRequest)
						w.Write(resp)
					}))
					hasuraClient, err := client.NewClient(nil, fmt.Sprintf("%s/", ts.URL))
					if err != nil {
						t.Fatal(err)
					}
					return hasuraClient
				}(),
				endpoint: "fail",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewCatalogState(tt.fields.client, tt.fields.endpoint)
			got, err := c.GetCatalogState()
			if !tt.wantErr {
				assert.Empty(t, err)
			} else {
				assert.NotEmpty(t, err.Error())
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCatalogState_SetCatalogState(t *testing.T) {
	type fields struct {
		client *client.Client
		route  string
	}
	type args struct {
		input *metadata.SetCatalogStateRequestArgs
	}

	m := http.NewServeMux()
	ts := httptest.NewServer(m)

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *metadata.SetCatalogStateResponseBody
		wantErr bool
	}{
		{
			name: "can set catalog state",
			fields: fields{
				client: func() *client.Client {
					m.Handle("/success", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						var resp = []byte(`
{
    "message": "success"
}
`)

						w.Header().Set("Content-Type", "application/json")
						w.Write(resp)
					}))
					hasuraClient, err := client.NewClient(nil, fmt.Sprintf("%s/", ts.URL))
					if err != nil {
						t.Fatal(err)
					}
					return hasuraClient
				}(),
				route: "success",
			},
			args: args{
				input: &metadata.SetCatalogStateRequestArgs{
					Type: metadata.CatalogStateCLIBackend,
					State: map[string]interface{}{
						"test": "test",
					},
				},
			},
			want: &metadata.SetCatalogStateResponseBody{
				Message: "success",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CatalogState{
				client: tt.fields.client,
				route:  tt.fields.route,
			}
			got, err := c.SetCatalogState(tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetCatalogState() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SetCatalogState() got = %v, want %v", got, tt.want)
			}
		})
	}
}
