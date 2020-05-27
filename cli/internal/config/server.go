package config

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/pkg/errors"
)

// ServerConfig has the config values required to contact the server
type ServerConfig struct {
	// Endpoint for the GraphQL Engine
	Endpoint string `yaml:"endpoint"`
	// AccessKey (deprecated) (optional) Admin secret key required to query the endpoint
	AccessKey string `yaml:"access_key,omitempty"`
	// AdminSecret (optional) Admin secret required to query the endpoint
	AdminSecret string `yaml:"admin_secret,omitempty"`
	// APIPaths (optional) API paths for server
	APIPaths *ServerAPIPaths `yaml:"api_paths,omitempty"`
	// InsecureSkipTLSVerify - indicates if TLS verification is disabled or not.
	InsecureSkipTLSVerify bool `yaml:"insecure_skip_tls_verify,omitempty"`
	// CAPath - Path to a cert file for the certificate authority
	CAPath string `yaml:"certificate_authority,omitempty"`

	ParsedEndpoint *url.URL `yaml:"-"`

	TLSConfig *tls.Config `yaml:"-"`

	HTTPClient *http.Client `yaml:"-"`
}

// ServerAPIPaths has the custom paths defined for server api
type ServerAPIPaths struct {
	Query   string `yaml:"query,omitempty"`
	GraphQL string `yaml:"graphql,omitempty"`
	Config  string `yaml:"config,omitempty"`
	PGDump  string `yaml:"pg_dump,omitempty"`
	Version string `yaml:"version,omitempty"`
}

// GetQueryParams - encodes the values in url
func (s ServerAPIPaths) GetQueryParams() url.Values {
	vals := url.Values{}
	t := reflect.TypeOf(s)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("yaml")
		splitTag := strings.Split(tag, ",")
		if len(splitTag) == 0 {
			continue
		}
		name := splitTag[0]
		if name == "-" {
			continue
		}
		v := reflect.ValueOf(s).Field(i)
		vals.Add(name, v.String())
	}
	return vals
}

// GetVersionEndpoint provides the url to contact the version API
func (s *ServerConfig) GetVersionEndpoint() string {
	nurl := *s.ParsedEndpoint
	nurl.Path = path.Join(nurl.Path, s.APIPaths.Version)
	return nurl.String()
}

// GetQueryEndpoint provides the url to contact the query API
func (s *ServerConfig) GetQueryEndpoint() string {
	nurl := *s.ParsedEndpoint
	nurl.Path = path.Join(nurl.Path, s.APIPaths.Query)
	return nurl.String()
}

// ParseEndpoint ensures the endpoint is valid.
func (s *ServerConfig) ParseEndpoint() error {
	nurl, err := url.ParseRequestURI(s.Endpoint)
	if err != nil {
		return err
	}
	s.ParsedEndpoint = nurl
	return nil
}

// SetTLSConfig - sets the TLS config
func (s *ServerConfig) SetTLSConfig() error {
	if s.InsecureSkipTLSVerify {
		s.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	}
	if s.CAPath != "" {
		// Get the SystemCertPool, continue with an empty pool on error
		rootCAs, _ := x509.SystemCertPool()
		if rootCAs == nil {
			rootCAs = x509.NewCertPool()
		}
		// read cert
		certPath, _ := filepath.Abs(s.CAPath)
		cert, err := ioutil.ReadFile(certPath)
		if err != nil {
			return errors.Errorf("error reading CA %s", s.CAPath)
		}
		if ok := rootCAs.AppendCertsFromPEM(cert); !ok {
			return errors.Errorf("Unable to append given CA cert.")
		}
		s.TLSConfig = &tls.Config{
			RootCAs:            rootCAs,
			InsecureSkipVerify: s.InsecureSkipTLSVerify,
		}
	}
	return nil
}

// SetHTTPClient - sets the http client
func (s *ServerConfig) SetHTTPClient() error {
	s.HTTPClient = &http.Client{Transport: http.DefaultTransport}
	if s.TLSConfig != nil {
		tr := &http.Transport{TLSClientConfig: s.TLSConfig}
		s.HTTPClient.Transport = tr
	}
	return nil
}
