package mockghauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type Codes struct {
	lock  sync.RWMutex
	codes map[string]time.Time
}

func (c *Codes) checkMap() {
	if c.codes != nil {
		return
	}

	c.codes = make(map[string]time.Time)
}

func (c *Codes) New() string {
	id := ulid.Make()

	c.lock.Lock()
	defer c.lock.Unlock()
	c.checkMap()

	c.codes[id.String()] = time.Now()

	return id.String()
}

func (c *Codes) ReadFile(filename string) error {
	if filename != "" {
		buf, fileErr := os.ReadFile(filename)
		if fileErr != nil {
			return fmt.Errorf("unable to read code-file(%s): %w", filename, fileErr)
		}

		if err := json.NewDecoder(bytes.NewReader(buf)).Decode(&c.codes); err != nil {
			return fmt.Errorf("unable to parse code-file(%s): %w", filename, err)
		}
	}

	return nil
}

func (c *Codes) WriteFile(filename string) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(c.codes); err != nil {
		return err
	}
	return os.WriteFile(filename, buf.Bytes(), 0o600)
}

func (c *Codes) Add(code string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.checkMap()
	c.codes[code] = time.Now()
}

func (c *Codes) Delete(code string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.checkMap()
	delete(c.codes, code)
}

func (c *Codes) Get(code string) (time.Time, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.checkMap()
	if v, ok := c.codes[code]; ok {
		return v, true
	}

	return time.Time{}, false
}

func (c *Codes) Exists(code string) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()

	c.checkMap()
	_, ok := c.codes[code]

	return ok
}
