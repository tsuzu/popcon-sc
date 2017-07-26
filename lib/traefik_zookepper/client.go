package traefikConsul

import (
	"net/url"
	"encoding/json"
	"time"
	"github.com/samuel/go-zookeeper/zk"
)

type Client struct {
	prefix string
	client *zk.Conn
}

func NewClient(prefix, addr string) (*Client, error) {
	conn, _, err := zk.Connect([]string{addr}, 30 * time.Second)

	return &Client{
		prefix: prefix,
		client: conn,
	}, nil
}

func (c *Client) RegisterNewBackend(backend, serverName, addr string) error {
	_, err := c.client.Create(
		c.prefix + "/backends/" + backend + "/servers/" + serverName + "/url",
		[]byte(addr),
		0, nil,
	)

	return err
}

func (c* Client) BackendGetAll(backend, serverName string) ([]string, error) {
	
}

func (c *Client) BackupBackend(backend, serverName string) ([]byte, error) {
	keyPrefix := c.prefix + "/backends/" + backend + "/servers/" + serverName + "/"
	pairs, _, err := c.client.Children(keyPrefix)

	if err != nil {
		return nil, err
	}

	m := make(map[string][]byte)
	for i := range pairs {
		m[pairs[i].Key[len(keyPrefix):]] = pairs[i].Value
	}
	b, _ := json.Marshal(m)

	return b, nil
}

func (c *Client) RestoreBackup(backend, serverName string, backup []byte) error {
	keyPrefix := c.prefix + "/backends/" + backend + "/servers/" + serverName + "/"
	var m map[string][]byte
	if err := json.Unmarshal(backup, &m); err != nil {
		return err
	}

	for k, v := range m {
		_, err := c.client.Create(
			keyPrefix + k,
			v,
			0, nil,
		)

		if err != nil {
			return err
		}
	}

	return nil
}

// Before executing DeleteBackend, you should execute BackupBackend
func (c *Client) DeleteBackend(backend, serverName string) error {
	_, err := c.client.KV().DeleteTree(c.prefix+"/backends/"+backend+"/servers/"+serverName, nil)

	return err
}

func (c *Client) NewFrontend(frontendName, backendName string) error {
	_, err := c.client.KV().Put(&api.KVPair{Key: c.prefix + "/frontends/" + frontendName + "/backend", Value: []byte(backendName)}, nil)

	return err
}

func (c *Client) HasFrontend() (bool, error) {
	pairs, _, err := c.client.KV().List(c.prefix+"/frontends/", nil)

	if err != nil {
		return false, err
	}

	return len(pairs) != 0, nil
}
