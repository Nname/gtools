package elasticsearch

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	elasticsearch7 "github.com/elastic/go-elasticsearch/v7"
	elasticsearch8 "github.com/elastic/go-elasticsearch/v8"
)

type AppService struct {
	Host    []string
	User    string
	Auth    string
	Timeout time.Duration
}

func (e *AppService) ClientV7() (*elasticsearch7.Client, error) {
	config := elasticsearch7.Config{
		Addresses: e.Host,
		Username:  e.User,
		Password:  e.Auth,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
	client, err := elasticsearch7.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (e *AppService) ClientV8() (*elasticsearch8.Client, error) {
	config := elasticsearch8.Config{
		Addresses: e.Host,
		Username:  e.User,
		Password:  e.Auth,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			DialContext:           (&net.Dialer{Timeout: time.Second}).DialContext,
			TLSClientConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
			},
		},
	}
	client, err := elasticsearch8.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (e *AppService) Info() {
}

func (e *AppService) IndexGetV7(index string) (interface{}, error) {
	v7, err := e.ClientV7()
	if err != nil {
		return nil, err
	}
	get, err := v7.Indices.Get([]string{index})
	if err != nil {
		return nil, err
	}
	return get.String(), nil
}
