package v1

import (
	"flag"
	"net/http"
	"net/url"
	"reflect"
	"testing"
)

func TestClientMetadataAndSchema_SendQuery(t *testing.T) {
	flag.Parse()
	if !*hasura {
		t.Skip()
	}
	type fields struct {
		SchemaAndMetadataAPIEndpoint url.URL
		Headers                      map[string]string
	}
	type args struct {
		m interface{}
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantResp *http.Response
		wantBody []byte
		wantErr  bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &ClientMetadataAndSchema{
				SchemaAndMetadataAPIEndpoint: tt.fields.SchemaAndMetadataAPIEndpoint,
				Headers:                      tt.fields.Headers,
			}
			gotResp, gotBody, err := client.SendQuery(tt.args.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientMetadataAndSchema.SendQuery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResp, tt.wantResp) {
				t.Errorf("ClientMetadataAndSchema.SendQuery() gotResp = %v, want %v", gotResp, tt.wantResp)
			}
			if !reflect.DeepEqual(gotBody, tt.wantBody) {
				t.Errorf("ClientMetadataAndSchema.SendQuery() gotBody = %v, want %v", gotBody, tt.wantBody)
			}
		})
	}
}
