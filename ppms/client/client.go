package ppms

import (
	"errors"
	"net/http"

	"net/url"

	"github.com/cs3238-tsuzu/popcon-sc/types"
)

type Client struct {
	addr, auth string
}

func (client *Client) defaultRequest() *http.Request {
	header := http.Header{}

	header.Set(sctypes.InternalHTTPToken, client.auth)

	return &http.Request{
		Method: "POST",
		Header: header,
	}
}

func (client *Client) RemoveFile(category, name string) error {
	req := client.defaultRequest()

	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}

	u.Path = "/remove_file"

	req.URL = u

	val := url.Values{}
	val.Add("category", category)
	val.Set("path", name)

	req.Form = val

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("error: " + resp.Status)
	}

	return nil
}

func NewClient(addr, auth string) (*Client, error) {
	return &Client{
		addr: addr,
		auth: auth,
	}, nil
}
