package app

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func ensureGoogleAPIKey(stdout io.Writer, stderr io.Writer) error {
	if strings.TrimSpace(os.Getenv("GEMINI_API_KEY")) != "" || strings.TrimSpace(os.Getenv("GOOGLE_API_KEY")) != "" {
		return nil
	}

	if !isInteractiveTerminal() {
		return fmt.Errorf("GEMINI_API_KEY (or GOOGLE_API_KEY) is required")
	}

	fmt.Fprintln(stdout, "Hello! Please enter your Google API key.")
	fmt.Fprint(stdout, "GEMINI_API_KEY: ")

	reader := bufio.NewReader(os.Stdin)
	value, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read API key: %w", err)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("empty API key provided")
	}

	if err := os.Setenv("GEMINI_API_KEY", value); err != nil {
		return fmt.Errorf("failed to set API key: %w", err)
	}

	fmt.Fprintln(stdout, "API key set for this session.")
	fmt.Fprintln(stdout, "Tip: export GEMINI_API_KEY in your shell profile to skip this prompt.")
	return nil
}

func isInteractiveTerminal() bool {
	info, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}
