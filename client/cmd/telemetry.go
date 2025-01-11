package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"alda.io/client/color"
	"alda.io/client/generated"
	"alda.io/client/help"
	log "alda.io/client/logging"
	"alda.io/client/system"
	"alda.io/client/text"
	"github.com/spf13/cobra"
)

var telemetryStatus bool
var telemetryEnable bool
var telemetryDisable bool

func init() {
	telemetryCmd.Flags().BoolVar(
		&telemetryStatus,
		"status",
		false,
		"Display whether telemetry is enabled or disabled",
	)

	telemetryCmd.Flags().BoolVar(
		&telemetryEnable,
		"enable",
		false,
		"Enable telemetry",
	)

	telemetryCmd.Flags().BoolVar(
		&telemetryDisable,
		"disable",
		false,
		"Disable telemetry",
	)
}

// TelemetryStatus describes the status of whether the user has been informed
// yet that we collect telemetry and whether they have chosen to disable
// telemetry.
type TelemetryStatus int

const (
	TelemetryNotInformed TelemetryStatus = iota
	TelemetryEnabled
	TelemetryDisabled
)

const telemetryStatusFileName = "telemetry-status"

func readTelemetryStatus() (TelemetryStatus, error) {
	// The environment variable ALDA_DISABLE_TELEMETRY can be set to "yes" as a
	// convenient way to disable telemetry for only one invocation of Alda, or to
	// disable telemetry only for the duration of a terminal session.
	if os.Getenv("ALDA_DISABLE_TELEMETRY") == "yes" {
		log.Info().
			Str("ALDA_DISABLE_TELEMETRY", "yes").
			Msg("Telemetry is disabled.")

		return TelemetryDisabled, nil
	}

	telemetryStatusFile := system.QueryConfig(telemetryStatusFileName)

	if telemetryStatusFile == "" {
		return TelemetryNotInformed, nil
	}

	content, err := os.ReadFile(telemetryStatusFile)
	if err != nil {
		return -1, err
	}

	switch str := string(content); str {
	case "enabled":
		return TelemetryEnabled, nil
	case "disabled":
		return TelemetryDisabled, nil
	default:
		return -1, fmt.Errorf("unrecognized telemetry file content: %s", str)
	}
}

func writeTelemetryStatus(status string) error {
	telemetryStatusFilepath := system.ConfigPath(telemetryStatusFileName)

	err := os.MkdirAll(filepath.Dir(telemetryStatusFilepath), os.ModePerm)
	if err != nil {
		return err
	}

	log.Debug().
		Str("status", status).
		Str("filepath", telemetryStatusFilepath).
		Msg("Writing telemetry status file.")

	return os.WriteFile(telemetryStatusFilepath, []byte(status), 0644)
}

func informUserOfTelemetry() {
	fmt.Fprintf(
		os.Stderr,
		text.Boxed(
			fmt.Sprintf(
				`%s

If you wish to disable anonymous usage reporting, you can run:

  %s`,
				telemetryExplanation,
				color.Aurora.BrightYellow("alda telemetry --disable"),
			),
		)+"\n\n",
	)

	if err := enableTelemetry(); err != nil {
		log.Warn().Err(err).Msg("Failed to enable telemetry.")
	}
}

func informUserOfTelemetryIfNeeded() {
	var status TelemetryStatus

	status, err := readTelemetryStatus()
	if err != nil {
		// If we can't read the telemetry status file for some reason, the only
		// reasonable thing we can do is assume they still need to be informed.
		status = TelemetryNotInformed
	}

	if status == TelemetryNotInformed {
		informUserOfTelemetry()
	}
}

func reportTelemetryStatus(status string) {
	fmt.Printf("Telemetry is %s.\n", color.Aurora.Bold(status))
}

func enableTelemetry() error {
	return writeTelemetryStatus("enabled")
}

func disableTelemetry() error {
	return writeTelemetryStatus("disabled")
}

func showTelemetryStatus() error {
	status, err := readTelemetryStatus()
	if err != nil {
		return err
	}

	switch status {
	case TelemetryNotInformed:
		informUserOfTelemetry()
		fmt.Println()
		reportTelemetryStatus("enabled")
	case TelemetryEnabled:
		reportTelemetryStatus("enabled")
	case TelemetryDisabled:
		reportTelemetryStatus("disabled")
	}

	return nil
}

const telemetryExplanation = `The Alda CLI collects the following anonymous usage statistics:

  • Operating system
  • Alda version
  • Command run, without arguments/options (e.g. alda play)

No personal information is collected.`

var telemetryCmd = &cobra.Command{
	Use:   "telemetry",
	Short: "Enable or disable telemetry",
	Long: fmt.Sprintf(`Enable or disable telemetry

---

%s

---`, telemetryExplanation),
	RunE: func(_ *cobra.Command, args []string) error {
		if telemetryEnable && telemetryDisable {
			return help.UserFacingErrorf(
				`%s and %s cannot be used together.

See %s for more information.`,
				color.Aurora.BrightYellow("--enable"),
				color.Aurora.BrightYellow("--disable"),
				color.Aurora.BrightYellow("alda telemetry --help"),
			)
		}

		if telemetryEnable {
			if err := enableTelemetry(); err != nil {
				return err
			}

			reportTelemetryStatus("enabled")
			return nil
		}

		if telemetryDisable {
			if err := disableTelemetry(); err != nil {
				return err
			}

			reportTelemetryStatus("disabled")
			return nil
		}

		return showTelemetryStatus()
	},
}

func aldaApiUrl(endpoint string) string {
	return "https://api.alda.io" + endpoint
}

func sendTelemetryRequest(command string) error {
	payload := map[string]string{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": generated.ClientVersion,
		"command": command,
	}

	json, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	log.Debug().Bytes("json", json).Msg("Sending telemetry request.")

	response, err := (&http.Client{Timeout: 5 * time.Second}).Post(
		aldaApiUrl("/telemetry"),
		"application/json",
		bytes.NewBuffer(json),
	)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	log.Debug().
		Int("status", response.StatusCode).
		Bytes("body", responseBody).
		Msg("Telemetry sent.")

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf(
			"unsuccessful response (%d): %s",
			response.StatusCode,
			responseBody,
		)
	}

	return nil
}
