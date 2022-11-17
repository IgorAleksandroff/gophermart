package webapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

//go:generate mockery --name "Client"

type (
	client struct {
		serverName string
		transport  *http.Client
	}

	Client interface {
		Do(req *http.Request) (body []byte, err error)
		DoGet(url string) ([]byte, error)
		DoPost(url string, data interface{}) ([]byte, error)
	}
)

func NewClient(serverName string) Client {

	return &client{
		serverName: serverName,
		transport:  &http.Client{},
	}
}

func (c client) Do(req *http.Request) (body []byte, err error) {
	r, err := c.transport.Do(req)
	fmt.Printf("for order %s status code: %v\n", req.URL.Path, r.StatusCode)
	if err != nil {
		return nil, fmt.Errorf("status code: %v; %w\n", r.StatusCode, err)
	}
	defer r.Body.Close()

	body, err = io.ReadAll(r.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func (c client) DoGet(path string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, c.serverName+path, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.Do(req)

	return body, err
}

func (c client) DoPost(path string, data interface{}) ([]byte, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		log.Println("payload marshal error")

		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, c.serverName+path, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set(`Content-Type`, `application/json`)

	body, err := c.Do(req)

	return body, err
}

var _ Client = &client{}
