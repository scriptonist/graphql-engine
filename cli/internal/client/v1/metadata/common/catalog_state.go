package common

import (
	"context"
	"net/http"

	"github.com/pkg/errors"

	"github.com/mitchellh/mapstructure"

	"github.com/hasura/graphql-engine/cli/internal/client"
	"github.com/hasura/graphql-engine/cli/internal/client/v1/metadata"
)

const (
	defaultCatalogStateRoute = "v1/metadata"
)

type CatalogState struct {
	client *client.Client
	// route at which CatalogState is available
	// default: v1/metadata
	route string
}

func NewCatalogState(client *client.Client, route string) *CatalogState {
	if route == "" {
		route = defaultCatalogStateRoute
	}
	c := new(CatalogState)
	c.client = client
	c.route = route
	return c
}

func (c *CatalogState) GetCatalogState() (*metadata.GetCatalogStateResponseBody, error) {
	body := new(metadata.GetCatalogStateRequest)
	req, err := metadata.NewMetadataHTTPRequest(c.client, c.route, "POST", body)
	if err != nil {
		return nil, err
	}

	var responseBody = new(map[string]interface{})
	resp, err := c.client.Do(context.Background(), req, responseBody)
	if resp.StatusCode == http.StatusBadRequest {
		var catalogStateErr = new(metadata.CatalogStateErr)
		if decodeErr := mapstructure.Decode(responseBody, catalogStateErr); decodeErr != nil {
			return nil, errors.Wrap(err, errors.Wrap(decodeErr, "decoding error body failed").Error())
		}
		if catalogStateErr != nil {
			return nil, catalogStateErr
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "HTTP request failed, code: %v body: %v", resp.StatusCode, responseBody)
	}

	var catalogStateOutput = new(metadata.GetCatalogStateResponseBody)
	if err := mapstructure.Decode(responseBody, catalogStateOutput); err != nil {
		return nil, err
	}

	return catalogStateOutput, nil
}

func (c *CatalogState) SetCatalogState(args *metadata.SetCatalogStateRequestArgs) (*metadata.SetCatalogStateResponseBody, error) {
	body := metadata.NewSetCatalogStateRequest(args)
	req, err := metadata.NewMetadataHTTPRequest(c.client, c.route, "POST", body)
	if err != nil {
		return nil, err
	}

	var responseBody = new(map[string]interface{})
	resp, err := c.client.Do(context.Background(), req, responseBody)
	if resp.StatusCode == http.StatusBadRequest {
		var catalogStateErr = new(metadata.CatalogStateErr)
		if decodeErr := mapstructure.Decode(responseBody, catalogStateErr); decodeErr != nil {
			return nil, errors.Wrap(err, errors.Wrap(decodeErr, "decoding error body failed").Error())
		}
		if catalogStateErr != nil {
			return nil, catalogStateErr
		}
	}
	if err != nil {
		return nil, errors.Wrapf(err, "HTTP request failed, code: %v body: %v", resp.StatusCode, responseBody)
	}

	var catalogStateOutput = new(metadata.SetCatalogStateResponseBody)
	if err := mapstructure.Decode(responseBody, catalogStateOutput); err != nil {
		return nil, err
	}

	return catalogStateOutput, nil
}
