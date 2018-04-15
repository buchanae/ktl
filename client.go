package ktl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// Client provides access to the ktl REST API.
type Client struct {
	Address string
	Client  *http.Client
}

// NewClient returns a new Client.
func NewClient(address string) *Client {
	return &Client{
		Address: address,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CreateBatch creates a new Batch and returns the ID of the created batch.
func (c *Client) CreateBatch(ctx context.Context, b *Batch) (*CreateBatchResponse, error) {
	err := ValidateBatch(b)
	if err != nil {
		return nil, fmt.Errorf("validating batch: %s", err)
	}

	by, err := json.Marshal(b)
	if err != nil {
		return nil, fmt.Errorf("marshaling batch: %s", err)
	}

	u := c.Address + "/v0/batch"
	req, err := http.NewRequest("POST", u, bytes.NewBuffer(by))
	if err != nil {
		return nil, fmt.Errorf("creating request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")

	out := &CreateBatchResponse{}
	err = c.do(ctx, req, out)
	return out, err
}

// ListBatches lists batches.
// TODO pass in list options.
func (c *Client) ListBatches(ctx context.Context) (*BatchList, error) {
	u := c.Address + "/v0/batch"
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")

	out := &BatchList{}
	err = c.do(ctx, req, out)
	return out, err
}

// GetBatch gets a batch by ID.
func (c *Client) GetBatch(ctx context.Context, id string) (*Batch, error) {
	u := c.Address + "/v0/batch/" + id
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %s", err)
	}
	req.Header.Add("Content-Type", "application/json")

	out := &Batch{}
	err = c.do(ctx, req, out)
	return out, err
}

func (c *Client) RestartStep(ctx context.Context, batchID, stepID string) error {
  u := c.Address + "/v0/batch/" + batchID +"/step/"+ stepID +":restart"
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %s", err)
	}
  return c.do(ctx, req, nil)
}

// do helps execute requests and deal with http responses; it checks for errors,
// reads the body, and unmarshals the response into the given "out" object.
func (c *Client) do(ctx context.Context, req *http.Request, out interface{}) error {
	req.WithContext(ctx)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	// shortcut: don't read the body if "out" is nil (caller doesn't want any unmarshaling)
	if resp.StatusCode == 200 && out == nil {
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading http response body: %s", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("http response %d: %s", resp.StatusCode, body)
	}

	err = json.Unmarshal(body, out)
	if err != nil {
		return fmt.Errorf("unmarshaling response: %s", err)
	}
	return nil
}
