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

// Client interface for client implementations.
type Client interface {
	Get(path string, v interface{}) error
	GetBytes(path string) ([]byte, error)

	Post(path string, v interface{}) error
	PostBytes(path string, data []byte) error

	Put(path string, v interface{}) error
	PutBytes(path string, data []byte) error

	Delete(path string) error
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

// GetBytes performs a get request and returns the response as a byte array.
func (c *HttpClient) GetBytes(path string) ([]byte, error) {
	url := c.createURL(path)
	req, err := newRequest(http.MethodGet, url, nil)
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

// PostBytes stores given data with path as a key.
func (c *HttpClient) PostBytes(path string, data []byte) error {
	req, err := newRequest(http.MethodPost, c.createURL(path), bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Non OK status: %d", resp.StatusCode)
	}
	var storeResp Response
	err = json.NewDecoder(resp.Body).Decode(&storeResp)
	if err != nil {
		return err
	}
	if !storeResp.OK {
		return fmt.Errorf("Failed to store resource at '%s'", path)
	}
	return nil
}

// Get gets the value for a given key and deserializes the result into a holder.
func (c *HttpClient) Get(path string, v interface{}) error {
	rawResponse, err := c.GetBytes(path)
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
		return fmt.Errorf("Could not get resource '%s'", path)
	}
	return resp.unmarshallResult(v)
}

func (c *HttpClient) createURL(resourcePath string) string {
	url := *c.baseURL
	url.Path = path.Join(url.Path, resourcePath)
	return url.String()
}

func (resp *Response) unmarshallResult(v interface{}) error {
	bytes, err := json.Marshal(resp.Result)
	if err != nil {
		return nil
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
	req.Header.Set("Content-type", "application/json")
	return req, nil
}
