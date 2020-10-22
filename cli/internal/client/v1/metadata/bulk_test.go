package metadata

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

func TestBulkQuery_Send(t *testing.T) {
	type fields struct {
		BulkQueryRequest  BulkQueryRequest
		BulkQueryHTTP     BulkQueryHTTP
		BulkQueryResponse *BulkQueryResponse
		ErrBulkQuery      *ErrBulkQuery
	}
	type args struct {
		client *client.Client
	}
	m := http.NewServeMux()
	ts := httptest.NewServer(m)
	defer ts.Close()

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"can send a bulk query",
			func() fields {
				r := BulkQueryRequest{
					args: []QueryType{
						func() QueryType {
							s := NewGetCatalogStateQuery("/catalog-state")
							q := NewQueryType(
								s.GetCatalogStateRequest,
								s.GetCatalogStateResponseBody,
								s.ErrGetCatalogStateQuery,
								s.GetCatalogStateHTTP,
							)
							return *q
						}(),
					},
				}
				bq := NewBulkQuery(r, "bulk")
				return fields{
					BulkQueryRequest:  bq.BulkQueryRequest,
					BulkQueryHTTP:     bq.BulkQueryHTTP,
					BulkQueryResponse: bq.BulkQueryResponse,
					ErrBulkQuery:      bq.ErrBulkQuery,
				}
			}(),
			args{
				client: func() *client.Client {
					m.HandleFunc("/bulk", func(w http.ResponseWriter, req *http.Request) {
						out := []byte(`
[{
    "id": "77d0a2cf-ea89-454a-9159-d24de958271b",
    "cli_state": {
        "key": "value"
    },
    "console_state": {}
}]
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
			b := BulkQuery{
				BulkQueryRequest:  tt.fields.BulkQueryRequest,
				BulkQueryHTTP:     tt.fields.BulkQueryHTTP,
				BulkQueryResponse: tt.fields.BulkQueryResponse,
				ErrBulkQuery:      tt.fields.ErrBulkQuery,
			}
			if err := b.Send(tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Println(tt.fields.BulkQueryResponse)
		})
	}
}
