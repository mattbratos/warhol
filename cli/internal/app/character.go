package app

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func runCharacter(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprintln(stderr, "missing character subcommand (expected: init)")
		return 2
	}

	switch args[0] {
	case "init":
		return runCharacterInit(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "unknown character subcommand: %s\n", args[0])
		return 2
	}
}

func runCharacterInit(args []string, stdout io.Writer, stderr io.Writer) int {
	fs := flag.NewFlagSet("character init", flag.ContinueOnError)
	fs.SetOutput(stderr)

	output := fs.String("output", "", "Path to output YAML file")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(stderr, "usage: warhol character init <name> [--output <path>]")
		return 2
	}

	name := rest[0]
	path := *output
	if path == "" {
		path = filepath.Join(defaultProjectPath("characters"), name+".yaml")
	}

	if err := writeCharacterTemplate(path, name); err != nil {
		fmt.Fprintf(stderr, "failed to write character template: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Created character template: %s\n", path)
	return 0
}

func writeCharacterTemplate(path string, characterName string) error {
	if _, err := os.Stat(path); err == nil {
		return errors.New("file already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	content := fmt.Sprintf(`# warhol character profile
name: %s
description: "Short character description."

traits:
  - "age range"
  - "hair style and color"

outfit:
  - "top clothing"
  - "bottom clothing"
  - "footwear"

prompt: ""
`, characterName)

	return os.WriteFile(path, []byte(content), 0o644)
}
