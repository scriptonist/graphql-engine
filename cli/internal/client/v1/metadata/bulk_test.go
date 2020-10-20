package metadata

import (
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
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"can send a bulk query",
			fields{
				BulkQueryRequest:  BulkQueryRequest{},
				BulkQueryHTTP:     BulkQueryHTTP{},
				BulkQueryResponse: nil,
				ErrBulkQuery:      nil,
			},
			args{},
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
		})
	}
}
