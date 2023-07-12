package mockghauth_test

import (
	"testing"
	"time"

	"github.com/dosquad/mock-oauth-test-server/mockghauth"
)

func TestTokens_ReadFile(t *testing.T) {
	testTime, _ := time.Parse(time.RFC3339, "2023-07-01T10:09:08.123456+10:00")

	tests := []struct {
		name     string
		fields   map[string]time.Time
		filename string
		wantErr  bool
	}{
		{
			name: "Reading Test Data",
			fields: map[string]time.Time{
				"test-code": testTime,
			},
			filename: "../testdata/tokens.json",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &mockghauth.Tokens{}

			if err := tr.ReadFile(tt.filename); (err != nil) != tt.wantErr {
				t.Errorf("Tokens.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for k, v := range tt.fields {
				if !tr.Exists(k) {
					t.Errorf("Tokens.ReadFile() key = %s, expected to exist", k)
					return
				}

				ts, ok := tr.GetExpire(k)
				if !ok {
					t.Errorf("Tokens.ReadFile() key = %s, expected to exist", k)
					return
				}

				if !v.Equal(ts) {
					t.Errorf("Tokens.ReadFile() key = %s, expected = %s, received = %s", k, v.String(), ts.String())
				}
			}
		})
	}
}

func TestTokens_Reaper(t *testing.T) {
	reaperTime := time.Now().Add(-1 * time.Second)
	tr := &mockghauth.Tokens{}
	tr.SetExpire(time.Second)

	expireToken := tr.New()

	if !tr.Exists(expireToken) {
		t.Errorf("Tokens.Reaper() key = %s, expected to exist", expireToken)
		return
	}

	tr.Reaper(reaperTime)

	if tr.Exists(expireToken) {
		t.Errorf("Tokens.Reaper() key = %s, expected to have been reaped", expireToken)
		return
	}
}
