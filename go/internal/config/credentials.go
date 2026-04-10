package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Host         string `yaml:"host"`
	APIKeyID     string `yaml:"api_key_id"`
	APIKeySecret string `yaml:"api_key_secret"`
	Scheme       string `yaml:"scheme,omitempty"` // "http" or "https" (default: "https")
}

// BaseURL returns the full base URL for the profile (e.g. "https://cloud.soda.io").
func (p Profile) BaseURL() string {
	host := p.Host
	if host == "" {
		host = "cloud.soda.io"
	}
	scheme := p.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return scheme + "://" + host
}

type Credentials map[string]Profile

func CredentialsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".soda", "credentials"), nil
}

func LoadCredentials() (Credentials, error) {
	path, err := CredentialsPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Credentials{}, nil
		}
		return nil, err
	}
	var creds Credentials
	if err := yaml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("invalid credentials file: %w", err)
	}
	return creds, nil
}

func SaveCredentials(creds Credentials) error {
	path, err := CredentialsPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := yaml.Marshal(creds)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func GetProfile(profileName string, creds Credentials) (Profile, error) {
	if profileName == "" {
		profileName = "default"
	}
	p, ok := creds[profileName]
	if !ok {
		return Profile{}, fmt.Errorf("profile '%s' not found in ~/.soda/credentials — run `sodacli auth login`", profileName)
	}
	if p.APIKeyID == "" || p.APIKeySecret == "" {
		return Profile{}, fmt.Errorf("profile '%s' is missing credentials — run `sodacli auth login`", profileName)
	}
	return p, nil
}
