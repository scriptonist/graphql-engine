package common

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

func Test_getCatalogState_GetCatalogState(t *testing.T) {
	type fields struct {
		path   string
		method string
	}
	type args struct {
		client  *client.Client
		request getCatalogStateRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *GetCatalogStateResponse
		wantErr bool
	}{
		{
			"can get current catalog state",
			fields{
				path:   "v1/metadata",
				method: http.MethodPost,
			},
			args{
				client: func() *client.Client {
					c, err := client.NewClient(nil, "http://localhost:8080/")
					if err != nil {
						t.Fatal(err)
					}
					return c
				}(),
				request: NewGetCatalogStateRequest(),
			},
			&GetCatalogStateResponse{
				ID:           "b934d747-b5c4-4f23-a47c-543c111fcfdb",
				CLIState:     func() *map[string]interface{} { m := map[string]interface{}{}; return &m }(),
				ConsoleState: func() *map[string]interface{} { m := map[string]interface{}{}; return &m }(),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gcs := &getCatalogState{
				path:   tt.fields.path,
				method: tt.fields.method,
			}
			got, err := gcs.GetCatalogState(tt.args.client, tt.args.request)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)

		})
	}
}
