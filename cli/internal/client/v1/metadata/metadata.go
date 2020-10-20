package metadata

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/hasura/graphql-engine/cli/internal/client"
)

const (
	defaultMetadataAPIPath = "v1/metadata"
)

// QueryType represents a query type and interfaces which should be implemented by
// a metadata query type
type QueryType struct {
	RequestBody   RequestBody
	ResponseBody  ResponseBody
	ResponseError ResponseError
	QueryHTTP     QueryHTTP
}

// Query is a helper interface which has to implemented by all query types
type HasuraMetadataV1Query interface {
	RequestBody
	ResponseBody
	ResponseError
	QueryHTTP
	Send(*client.Client) error
}

func NewQueryType(body RequestBody, response ResponseBody, error ResponseError, queryHttp QueryHTTP) *QueryType {
	return &QueryType{
		body,
		response,
		error,
		queryHttp,
	}
}

type QueryHTTP interface {
	Method() string
	// Path is excluding base url ie if metadata API is avialable at http://localhost:8080/v1/metadata
	// Path will be "v1/metadata"
	Path() string
}

// RequestBody is the interface which has to be satisfied by any type
// which is supposed to be a metadata request type supported by hasura
type RequestBody interface {
	Type() string
	Args() interface{}
}

type ResponseBody interface {
	UnmarshalResponseBodyJSON([]byte) error
}
type ResponseError interface {
	UnmarshalResponseErrorJSON([]byte) error
	error
}

func NewMetadataHTTPRequest(c *client.Client, path, method string, body RequestBody) (*http.Request, error) {
	var requestBody = struct {
		Type string
		Args interface{}
	}{
		body.Type(),
		body.Args(),
	}
	req, err := c.NewRequest(method, path, requestBody)
	if err != nil {
		return nil, err
	}

	return req, nil
}

// SendMetadataQuery takes an HTTP Client and a QueryType to send a metadata query request
// If the request succeeds with a 200
// the result is written to QueryType.ResponseBody
// If the request fails with a 400 bad request
// The error is written to QueryType.ResponseError
func SendMetadataQuery(client *client.Client, queryType *QueryType) error {
	req, err := NewMetadataHTTPRequest(client, queryType.QueryHTTP.Path(), queryType.QueryHTTP.Method(), queryType.RequestBody)
	if err != nil {
		return err
	}
	var responseBody = new(bytes.Buffer)
	response, err := client.Do(context.Background(), req, responseBody)
	if err != nil {
		return err
	}
	fmt.Println(responseBody.String())
	if response.StatusCode == http.StatusBadRequest {
		if err := queryType.ResponseError.UnmarshalResponseErrorJSON(responseBody.Bytes()); err != nil {
			return err
		}
		return nil
	} else if response.StatusCode != http.StatusOK {
		return err
	}

	if err := queryType.ResponseBody.UnmarshalResponseBodyJSON(responseBody.Bytes()); err != nil {
		return err
	}
	return nil
}
