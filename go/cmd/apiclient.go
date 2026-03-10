package cmd

import (
	"github.com/soda-data-inc/soda-cli/internal/api"
	"github.com/soda-data-inc/soda-cli/internal/config"
	"github.com/soda-data-inc/soda-cli/internal/output"
)

// newAPIClient builds an authenticated API client from the active profile.
// Returns exit code 3 if credentials are missing or invalid.
func newAPIClient() (*api.Client, error) {
	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, output.Errorf(2, "could not read credentials: %v", err)
	}
	profile, err := config.GetProfile(GCtx.Profile, creds)
	if err != nil {
		return nil, output.Errorf(3, "%v", err)
	}
	return api.New(profile), nil
}
