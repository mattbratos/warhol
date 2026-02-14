package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func runStyle(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "missing style subcommand (expected: init)")
		return 2
	}

	switch args[0] {
	case "init":
		return runStyleInit(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown style subcommand: %s\n", args[0])
		return 2
	}
}

func runStyleInit(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("style init", flag.ContinueOnError)
	fs.SetOutput(stderr)

	output := fs.String("output", "", "Path to output YAML file")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(stderr, "usage: warhol style init <name> [--output <path>]")
		return 2
	}

	name := rest[0]
	path := *output
	if path == "" {
		path = filepath.Join(defaultProjectPath("styles"), name+".yaml")
	}

	if err := writeStyleTemplate(path, name); err != nil {
		fmt.Fprintf(stderr, "failed to write style template: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Created style template: %s\n", path)
	return 0
}

func writeStyleTemplate(path string, styleName string) error {
	if _, err := os.Stat(path); err == nil {
		return errors.New("file already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content := fmt.Sprintf(`# warhol style profile
name: %s
description: "Short description of the intended visual identity."

prompt_prefix:
  - "Define the core visual style in plain language."

palette:
  - "#111111"
  - "#f4f4f4"

camera:
  lens: "35mm"
  framing: "medium shot"
  lighting: "soft directional light"

negative_prompt:
  - "avoid brand marks"
  - "avoid unrelated text"

seed_policy:
  mode: "fixed" # fixed | random
  seed: 42
`, styleName)

	return os.WriteFile(path, []byte(content), 0o644)
}
