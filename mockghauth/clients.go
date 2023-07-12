package mockghauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/oklog/ulid/v2"
)

type Client struct {
	ID     string `json:"id"`
	Secret string `json:"secret"`
}

func NewClient(id, secret string) *Client {
	return &Client{
		ID:     id,
		Secret: secret,
	}
}

func (c *Client) Equal(v *Client) bool {
	if v == nil {
		return false
	}

	return strings.EqualFold(c.ID, v.ID) && strings.EqualFold(c.Secret, v.Secret)
}

func (c *Client) String() string {
	return fmt.Sprintf("ID:%s, len(Secret):%d", c.ID, len(c.Secret))
}

type Clients struct {
	lock    sync.RWMutex
	clients map[string]*Client
}

func NewClients() *Clients {
	return &Clients{
		clients: make(map[string]*Client),
	}
}

func (c *Clients) ReadFile(filename string) error {
	if filename != "" {
		buf, fileErr := os.ReadFile(filename)
		if fileErr != nil {
			return fmt.Errorf("unable to read clients-file(%s): %w", filename, fileErr)
		}

		if err := json.NewDecoder(bytes.NewReader(buf)).Decode(&c.clients); err != nil {
			return fmt.Errorf("unable to parse clients-file(%s): %w", filename, err)
		}
	}

	return nil
}

func (c *Clients) WriteFile(filename string) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(c.clients); err != nil {
		return err
	}
	return os.WriteFile(filename, buf.Bytes(), 0o600)
}

func (c *Clients) New() (string, string) {
	uid := ulid.Make()
	token := "secret-token"

	c.lock.Lock()
	defer c.lock.Unlock()

	c.clients[uid.String()] = NewClient(uid.String(), token)

	return uid.String(), token
}

func (c *Clients) Add(id, secret string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.clients[id] = NewClient(id, secret)
}

func (c *Clients) HasID(id string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_, ok := c.clients[id]
	return ok
}

func (c *Clients) Get(id string) (*Client, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	v, ok := c.clients[id]
	return v, ok
}
