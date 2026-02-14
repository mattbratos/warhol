package app

import (
	"fmt"
	"io"
)

func Run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		return runWelcome(stdout, stderr)
	}

	switch args[0] {
	case "help", "--help", "-h":
		printUsage(stdout)
		return 0
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, version)
		return 0
	case "style":
		return runStyle(args[1:], stdout, stderr)
	case "character":
		return runCharacter(args[1:], stdout, stderr)
	case "generate":
		return runGenerate(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown command: %s\n\n", args[0])
		printUsage(stderr)
		return 2
	}
}

func printUsage(w io.Writer) {
	fmt.Fprintln(w, "warhol - CLI for creating images in a consistent visual style")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Usage:")
	fmt.Fprintln(w, "  warhol style init <name> [--output <path>]")
	fmt.Fprintln(w, "  warhol character init <name> [--output <path>]")
	fmt.Fprintln(w, "  warhol generate --style <path-or-name> [--character <name-or-path>|-<name>] --prompt <text> [--provider google|openai] [--out-dir <dir>]")
	fmt.Fprintln(w, "  warhol version")
}
