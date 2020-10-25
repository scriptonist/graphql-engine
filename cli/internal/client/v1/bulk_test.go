package v1

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hasura/graphql-engine/cli/internal/client/v1/common"
	"github.com/stretchr/testify/assert"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

func Test_bulk_Send(t *testing.T) {
	type fields struct {
		path   string
		method string
	}
	type args struct {
		client  *client.Client
		request bulkRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *bulkResponse
		wantErr bool
	}{
		{
			"can send bulk api request",
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
				request: NewBulkRequest(BulkRequestArgs{
					common.NewGetCatalogStateRequest(),
					common.NewGetCatalogStateRequest(),
				}),
			},
			&bulkResponse{
				common.GetCatalogStateResponse{
					ID:           "b934d747-b5c4-4f23-a47c-543c111fcfdb",
					CLIState:     func() *map[string]interface{} { m := map[string]interface{}{}; return &m }(),
					ConsoleState: func() *map[string]interface{} { m := map[string]interface{}{}; return &m }(),
				},
				common.GetCatalogStateResponse{
					ID:           "b934d747-b5c4-4f23-a47c-543c111fcfdb",
					CLIState:     func() *map[string]interface{} { m := map[string]interface{}{}; return &m }(),
					ConsoleState: func() *map[string]interface{} { m := map[string]interface{}{}; return &m }(),
				},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &bulk{
				path:   tt.fields.path,
				method: tt.fields.method,
			}
			got, err := b.Send(tt.args.client, tt.args.request)
			assert.NoError(t, err)
			gotb, err := json.Marshal(got)
			assert.NoError(t, err)
			wantb, err := json.Marshal(tt.want)
			assert.NoError(t, err)
			assert.JSONEq(t, string(wantb), string(gotb))
		})
	}
}
