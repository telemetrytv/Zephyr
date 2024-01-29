package zephyr

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/RobertWHurst/navaros"
)

type Client struct {
	Connection Connection
}

type ServiceClient struct {
	*Client
	Name string
}

func (c *Client) Service(name string) *ServiceClient {
	return &ServiceClient{
		Client: c,
		Name:   name,
	}
}

func (c *ServiceClient) Do(req *http.Request) (*http.Response, error) {
	responseRecorder := httptest.NewRecorder()
	if err := c.Connection.Dispatch(c.Name, responseRecorder, req); err != nil {
		return nil, err
	}
	return responseRecorder.Result(), nil
}

func (c *ServiceClient) Get(servicePath string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, servicePath, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *ServiceClient) Head(servicePath string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodHead, servicePath, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *ServiceClient) Post(servicePath string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, servicePath, body)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *ServiceClient) PostForm(servicePath string, data url.Values) (*http.Response, error) {
	return c.Post(servicePath, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (c *ServiceClient) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c.Connection.Dispatch(c.Name, w, r)
}

func (c *ServiceClient) Handle(ctx *navaros.Context) {
	c.ServeHTTP(ctx.ResponseWriter(), ctx.Request())
}
