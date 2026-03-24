package sodacore

import (
	"fmt"
	"os"
)

// WriteTempCloudConfig creates a temporary soda-cloud YAML config file
// from the given credentials. Returns the file path and a cleanup function.
func WriteTempCloudConfig(host, apiKeyID, apiKeySecret string) (string, func(), error) {
	content := fmt.Sprintf(`soda_cloud:
  host: %s
  api_key_id: %s
  api_key_secret: %s
`, host, apiKeyID, apiKeySecret)

	f, err := os.CreateTemp("", "soda-cloud-*.yml")
	if err != nil {
		return "", nil, fmt.Errorf("could not create temp file: %w", err)
	}

	if err := os.Chmod(f.Name(), 0600); err != nil {
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("could not set permissions: %w", err)
	}

	if _, err := f.WriteString(content); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, fmt.Errorf("could not write config: %w", err)
	}
	f.Close()

	cleanup := func() { os.Remove(f.Name()) }
	return f.Name(), cleanup, nil
}
