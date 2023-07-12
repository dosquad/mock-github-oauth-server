package mockghauth_test

import (
	"testing"
	"time"

	"github.com/dosquad/mock-oauth-test-server/mockghauth"
)

func TestCodes_ReadFile(t *testing.T) {
	testTime, _ := time.Parse(time.RFC3339, "2023-07-01T10:09:08.123456+10:00")

	tests := []struct {
		name     string
		fields   map[string]time.Time
		filename string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name: "Reading Test Data",
			fields: map[string]time.Time{
				"test-code": testTime,
			},
			filename: "../testdata/codes.json",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &mockghauth.Codes{}

			if err := c.ReadFile(tt.filename); (err != nil) != tt.wantErr {
				t.Errorf("Codes.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for k, v := range tt.fields {
				ts, ok := c.Get(k)
				if !ok {
					t.Errorf("Codes.ReadFile() key = %s, expected to exist", k)
					return
				}

				if !v.Equal(ts) {
					t.Errorf("Codes.ReadFile() key = %s, expected = %s, received = %s", k, v.String(), ts.String())
				}
			}
		})
	}
}
