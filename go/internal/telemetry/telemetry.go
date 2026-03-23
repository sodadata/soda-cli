// Package telemetry sends anonymous usage events to help improve the CLI.
// Telemetry is opt-out: set SODACLI_TELEMETRY=false or run `sodacli config set telemetry false`.
package telemetry

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

const endpoint = "https://ccnsnrtbwdhdjsloxwqb.supabase.co/functions/v1/cli-event"

type Event struct {
	MachineID  string   `json:"machine_id"`
	Command    string   `json:"command"`
	Flags      []string `json:"flags"`
	ExitCode   int      `json:"exit_code"`
	DurationMs int64    `json:"duration_ms"`
	CLIVersion string   `json:"cli_version"`
	OS         string   `json:"os"`
	Arch       string   `json:"arch"`
}

// Enabled returns false if the user has opted out of telemetry.
func Enabled() bool {
	v := os.Getenv("SODACLI_TELEMETRY")
	if v == "" {
		return true
	}
	return v != "false" && v != "0" && v != "off" && v != "no"
}

// Send fires a telemetry event in a goroutine and returns a function that
// waits for it to complete (with a timeout). Call the returned function
// before os.Exit to ensure the event is delivered.
func Send(event Event) (wait func()) {
	if !Enabled() {
		return func() {}
	}

	event.MachineID = machineID()
	event.OS = runtime.GOOS
	event.Arch = runtime.GOARCH

	done := make(chan struct{})
	go func() {
		defer close(done)
		data, err := json.Marshal(event)
		if err != nil {
			return
		}

		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Post(endpoint, "application/json", bytes.NewReader(data))
		if err != nil {
			return
		}
		resp.Body.Close()
	}()

	return func() {
		select {
		case <-done:
		case <-time.After(3 * time.Second):
		}
	}
}

// machineID returns a stable anonymous identifier derived from hostname + username.
// It is a one-way hash — the original values cannot be recovered.
func machineID() string {
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME")
	}
	h := sha256.Sum256([]byte(hostname + ":" + username))
	return fmt.Sprintf("%x", h[:8]) // 16-char hex string
}

// CommandName extracts the resource + action from os.Args (e.g., "dataset list").
// Returns the first two non-flag arguments after the binary name.
func CommandName(args []string) string {
	var parts []string
	for _, a := range args[1:] { // skip binary name
		if strings.HasPrefix(a, "-") {
			break
		}
		parts = append(parts, a)
		if len(parts) == 2 {
			break
		}
	}
	return strings.Join(parts, " ")
}

// FlagNames extracts flag names (without values) from os.Args.
// Never captures flag values — only the flag names like "--output", "--limit".
func FlagNames(args []string) []string {
	var flags []string
	for _, a := range args[1:] {
		if strings.HasPrefix(a, "--") {
			name := strings.SplitN(a, "=", 2)[0]
			flags = append(flags, name)
		} else if strings.HasPrefix(a, "-") && len(a) == 2 {
			flags = append(flags, a)
		}
	}
	return flags
}
