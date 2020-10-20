package metadata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

func TestGetCatalogStateQuery_Send(t *testing.T) {
	type fields struct {
		GetCatalogStateRequest      GetCatalogStateRequest
		GetCatalogStateHTTP         GetCatalogStateHTTP
		GetCatalogStateResponseBody *GetCatalogStateResponseBody
		ErrGetCatalogStateQuery     *ErrGetCatalogStateQuery
	}
	m := http.NewServeMux()
	ts := httptest.NewServer(m)
	type args struct {
		client *client.Client
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"can send catalog state request",
			func() fields {
				q := NewGetCatalogStateQuery("/success")
				return fields{
					GetCatalogStateRequest:      q.GetCatalogStateRequest,
					GetCatalogStateHTTP:         q.GetCatalogStateHTTP,
					GetCatalogStateResponseBody: q.GetCatalogStateResponseBody,
					ErrGetCatalogStateQuery:     q.ErrGetCatalogStateQuery,
				}
			}(),
			args{
				client: func() *client.Client {
					m.HandleFunc("/success", func(w http.ResponseWriter, req *http.Request) {
						out := []byte(`
{
    "id": "77d0a2cf-ea89-454a-9159-d24de958271b",
    "cli_state": {
        "key": "value"
    },
    "console_state": {}
}
`)
						w.Header().Set("Content-Type", "application/json")
						if _, err := w.Write(out); err != nil {
							t.Fatal(err)
						}
					})
					c, err := client.NewClient(nil, fmt.Sprintf("%s/", ts.URL))
					if err != nil {
						t.Fatal(err)
					}
					return c
				}(),
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			getCatalogStateQuery := &GetCatalogStateQuery{
				GetCatalogStateRequest:      tt.fields.GetCatalogStateRequest,
				GetCatalogStateHTTP:         tt.fields.GetCatalogStateHTTP,
				GetCatalogStateResponseBody: tt.fields.GetCatalogStateResponseBody,
				ErrGetCatalogStateQuery:     tt.fields.ErrGetCatalogStateQuery,
			}
			if err := getCatalogStateQuery.Send(tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
