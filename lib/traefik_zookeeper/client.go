package traefikZookeeper

import (
	"encoding/json"
	"github.com/docker/libkv/store"
	"github.com/docker/libkv/store/zookeeper"
)

type Client struct {
	prefix string
	client store.Store
}

func NewClient(prefix, addr string) (*Client, error) {
	store, err := zookeeper.New([]string{addr}, nil)

	if err != nil {
		return nil, err
	}

	return &Client{
		prefix: prefix,
		client: store,
	}, nil
}

func (c *Client) RegisterNewBackend(backend, serverName, addr string) error {
	err := c.client.Put(
		c.prefix+"/backends/"+backend+"/servers/"+serverName+"/url",
		[]byte(addr),
		nil,
	)

	return err
}

func (c *Client) BackupBackend(backend, serverName string) ([]byte, error) {
	keyPrefix := c.prefix + "/backends/" + backend + "/servers/" + serverName + "/"
	pairs, err := c.client.List(keyPrefix)

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
		if err := c.client.Put(
			keyPrefix+k,
			v,
			nil,
		); err != nil {
			return err
		}
	}

	return nil
}

// Before executing DeleteBackend, you should execute BackupBackend
func (c *Client) DeleteBackend(backend, serverName string) error {
	return c.client.DeleteTree(c.prefix + "/backends/" + backend + "/servers/" + serverName)
}

func (c *Client) NewFrontend(frontendName, backendName string) error {
	return c.client.Put(c.prefix+"/frontends/"+frontendName+"/backend", []byte(backendName), nil)
}

func (c *Client) HasFrontend() (bool, error) {
	pairs, err := c.client.List(c.prefix + "/frontends/")

	if err != nil {
		if err == store.ErrKeyNotFound {
			err = nil
		}
		return false, err
	}

	return len(pairs) != 0, nil
}
