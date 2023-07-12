package mockghauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

type Tokens struct {
	lock   sync.RWMutex
	expire time.Duration
	tokens map[string]time.Time
}

func (t *Tokens) checkMap() {
	if t.expire == 0 {
		t.expire = time.Minute
	}

	if t.tokens != nil {
		return
	}

	t.tokens = make(map[string]time.Time)
}

func (t *Tokens) SetExpire(exp time.Duration) {
	t.expire = exp
}

func (t *Tokens) ReadFile(filename string) error {
	if filename != "" {
		buf, fileErr := os.ReadFile(filename)
		if fileErr != nil {
			return fmt.Errorf("unable to read token-file(%s): %w", filename, fileErr)
		}

		if err := json.NewDecoder(bytes.NewReader(buf)).Decode(&t.tokens); err != nil {
			return fmt.Errorf("unable to parse token-file(%s): %w", filename, err)
		}
	}

	return nil
}

func (t *Tokens) WriteFile(filename string) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(t.tokens); err != nil {
		return err
	}
	return os.WriteFile(filename, buf.Bytes(), 0o600)
}

func (t *Tokens) New() string {
	t.lock.Lock()
	defer t.lock.Unlock()

	id := ulid.Make()

	token := fmt.Sprintf("ght_%s", strings.ToLower(id.String()))

	t.checkMap()
	t.tokens[token] = time.Now().Add(t.expire)

	return token
}

func (t *Tokens) Delete(token string) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.checkMap()
	delete(t.tokens, token)
}

func (t *Tokens) GetExpire(token string) (time.Time, bool) {
	t.lock.RLock()
	defer t.lock.RUnlock()

	t.checkMap()
	if v, ok := t.tokens[token]; ok {
		return v, true
	}

	return time.Time{}, false
}

func (t *Tokens) Exists(token string) bool {
	t.lock.RLock()
	defer t.lock.RUnlock()

	t.checkMap()
	_, ok := t.tokens[token]

	return ok
}

func (t *Tokens) Reaper(ts time.Time) {
	t.lock.Lock()
	defer t.lock.Unlock()

	t.checkMap()

	for k := range t.tokens {
		if t.tokens[k].After(ts) {
			delete(t.tokens, k)
		}
	}
}
