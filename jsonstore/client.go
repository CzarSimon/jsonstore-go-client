package jsonstore

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
)

var (
	JsonstoreUrl, _ = url.Parse("https://www.jsonstore.io")
	ErrNoValue      = errors.New("No value for key")
)

// Client interface for jsonstore client implementations.
type Client interface {
	Get(key string, v interface{}) error // Done
	GetBytes(key string) ([]byte, error) // Done

	Post(key string, v interface{}) error    // Done
	PostBytes(key string, data []byte) error // Done

	Put(key string, v interface{}) error
	PutBytes(key string, data []byte) error

	Delete(key string) error
}

// HttpClient main http client for interacting with jsonstore.
type HttpClient struct {
	httpClient *http.Client
	baseURL    *url.URL
}

// Response structure of responses returned from jsonstore.
type Response struct {
	Result interface{} `json:"result"`
	OK     bool        `json:"ok"`
}

// NewClient creates a new HttpClient.
func NewClient(storeKey string) *HttpClient {
	url := JsonstoreUrl
	url.Path = storeKey
	return &HttpClient{
		httpClient: createNetHttpClient(),
		baseURL:    url,
	}
}

// Get gets value from jsonstore.
func (c *HttpClient) Get(key string, v interface{}) error {
	rawResponse, err := c.GetBytes(key)
	if err != nil {
		return err
	}
	resp, err := newResponse(rawResponse)
	if err != nil {
		return err
	}
	if resp.Result == nil {
		return ErrNoValue
	}
	if !resp.OK {
		return fmt.Errorf("Could not get resource '%s'", key)
	}
	return resp.unmarshallResult(v)
}

// GetBytes gets value from jsonstore as a bytes.
func (c *HttpClient) GetBytes(key string) ([]byte, error) {
	req, err := newRequest(http.MethodGet, c.createURL(key), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Non OK status: %d", resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

// Post posts a value in jsonstore.
func (c *HttpClient) Post(key string, v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.PostBytes(key, body)
}

// PostBytes posts raw bytes to jsonstore.
func (c *HttpClient) PostBytes(key string, data []byte) error {
	req, err := newRequest(http.MethodPost, c.createURL(key), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = c.performRequest(key, req)
	return err
}

// Put updates the value of a given key in jsonstore.
func (c *HttpClient) Put(key string, v interface{}) error {
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return c.PutBytes(key, body)
}

// PutBytes updates the value of a given key in jsonstore.
func (c *HttpClient) PutBytes(key string, data []byte) error {
	req, err := newRequest(http.MethodPut, c.createURL(key), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	_, err = c.performRequest(key, req)
	return err
}

// Delete deletes the value of a key in jsonstore.
func (c *HttpClient) Delete(key string) error {
	req, err := newRequest(http.MethodDelete, c.createURL(key), nil)
	if err != nil {
		return err
	}
	_, err = c.performRequest(key, req)
	return err
}

func (c *HttpClient) performRequest(key string, r *http.Request) (*Response, error) {
	resp, err := c.httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Non OK status: %d", resp.StatusCode)
	}
	var storeResp Response
	err = json.NewDecoder(resp.Body).Decode(&storeResp)
	if err != nil {
		return nil, err
	}
	if !storeResp.OK {
		return nil, fmt.Errorf("Failed to store resource at '%s'", key)
	}
	return &storeResp, nil
}

func (c *HttpClient) createURL(resourcePath string) string {
	url := *c.baseURL
	url.Path = path.Join(url.Path, resourcePath)
	return url.String()
}

func (resp *Response) unmarshallResult(v interface{}) error {
	bytes, err := json.Marshal(resp.Result)
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, v)
}

func newResponse(data []byte) (*Response, error) {
	var resp Response
	err := json.Unmarshal(data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func createNetHttpClient() *http.Client {
	const TIMEOUT_SECONDS = 5
	return &http.Client{
		Timeout: TIMEOUT_SECONDS * time.Second,
	}
}

func newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-type", "application/json")
	return req, nil
}
