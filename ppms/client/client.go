package ppms

import (
	"errors"
	"net/http"

	"net/url"

	"strings"

	"github.com/cs3238-tsuzu/popcon-sc/types"
)

type Client struct {
	addr, auth string
}

func (client *Client) RemoveFile(category, name string) error {
	u, err := url.Parse(client.addr)

	if err != nil {
		return err
	}
	u.Path = "/remove_file"

	val := url.Values{}
	val.Add("category", category)
	val.Add("path", name)

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(val.Encode()))

	if err != nil {
		return err
	}

	req.Header.Set(sctypes.InternalHTTPToken, client.auth)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
