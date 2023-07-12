package mockghauth_test

import (
	"testing"

	"github.com/dosquad/mock-oauth-test-server/mockghauth"
)

func TestClients_ReadFile(t *testing.T) {
	tests := []struct {
		name     string
		fields   map[string]*mockghauth.Client
		filename string
		wantErr  bool
	}{
		// TODO: Add test cases.
		{
			name: "Reading Test Data",
			fields: map[string]*mockghauth.Client{
				"test-client": mockghauth.NewClient("test-client", "secret"),
			},
			filename: "../testdata/clients.json",
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := mockghauth.NewClients()

			if err := c.ReadFile(tt.filename); (err != nil) != tt.wantErr {
				t.Errorf("Clients.ReadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			for k, v := range tt.fields {
				cl, ok := c.Get(k)
				if !ok {
					t.Errorf("Clients.ReadFile() key = %s, expected to exist", k)
					return
				}

				if cl == nil {
					t.Errorf("Clients.ReadFile() key = %s, expected = %s, received = nil", k, v.String())
					return
				}

				if !v.Equal(cl) {
					t.Errorf("Clients.ReadFile() key = %s, expected = %s, received = %s", k, v.String(), cl.String())
				}
			}
		})
	}
}
