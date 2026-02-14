package app

import (
	"fmt"
	"io"
)

func runWelcome(stdout io.Writer, stderr io.Writer) int {
	fmt.Fprintln(stdout, "Hello! Welcome to warhol.")
	fmt.Fprintln(stdout, "I can generate images in a consistent style, and I need your Google API key first.")

	if err := ensureGoogleAPIKey(stdout, stderr); err != nil {
		fmt.Fprintf(stderr, "setup failed: %v\n", err)
		return 1
	}

	fmt.Fprintln(stdout)
	fmt.Fprintln(stdout, "You're ready. Try this next:")
	fmt.Fprintln(stdout, `warhol generate --style 16bit -matt --prompt "full body portrait, city street at night"`)
	return 0
}
