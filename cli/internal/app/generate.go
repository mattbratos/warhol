package app

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type generationManifest struct {
	CreatedAt     string `json:"created_at"`
	Provider      string `json:"provider"`
	Model         string `json:"model"`
	Size          string `json:"size,omitempty"`
	Quality       string `json:"quality,omitempty"`
	StyleInput    string `json:"style_input"`
	StyleFile     string `json:"style_file"`
	Character     string `json:"character,omitempty"`
	CharacterFile string `json:"character_file,omitempty"`
	Prompt        string `json:"prompt"`
	FinalPrompt   string `json:"final_prompt"`
	ImagePath     string `json:"image_path,omitempty"`
	DryRun        bool   `json:"dry_run"`
}

func runGenerate(args []string, stdout io.Writer, stderr io.Writer) int {
	args = normalizeGenerateArgs(args)

	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(stderr)

	style := fs.String("style", "", "Style profile path or name")
	character := fs.String("character", "", "Character profile path or name")
	prompt := fs.String("prompt", "", "Prompt text")
	outDir := fs.String("out-dir", defaultProjectPath("outputs"), "Directory for generated artifacts")
	provider := fs.String("provider", "google", "Image provider (google|openai)")
	model := fs.String("model", "", "Model override (defaults by provider)")
	size := fs.String("size", "1024x1024", "OpenAI image size (e.g. 1024x1024)")
	quality := fs.String("quality", "medium", "OpenAI image quality (e.g. low, medium, high)")
	dryRun := fs.Bool("dry-run", false, "Compose prompt and write metadata without generating image")

	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *style == "" || *prompt == "" {
		fmt.Fprintln(stderr, "usage: warhol generate --style <path-or-name> [--character <name-or-path>|-<name>] --prompt <text> [--provider google|openai] [--model <name>] [--out-dir <dir>]")
		return 2
	}

	ts := time.Now().UTC().Format("20060102-150405")
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintf(stderr, "failed to create output directory: %v\n", err)
		return 1
	}

	styleProfile, stylePath, err := loadStyleProfile(*style)
	if err != nil {
		fmt.Fprintf(stderr, "failed to load style profile: %v\n", err)
		return 1
	}

	var characterProfileData *characterProfile
	characterPath := ""
	if *character != "" {
		loadedCharacter, resolvedPath, err := loadCharacterProfile(*character)
		if err != nil {
			fmt.Fprintf(stderr, "failed to load character profile: %v\n", err)
			return 1
		}
		characterProfileData = &loadedCharacter
		characterPath = resolvedPath
	}

	finalPrompt := buildFinalPrompt(styleProfile, characterProfileData, *prompt)
	providerValue := strings.ToLower(*provider)

	resolvedModel, err := resolveModel(providerValue, *model)
	if err != nil {
		fmt.Fprintf(stderr, "invalid model/provider: %v\n", err)
		return 2
	}

	manifest := generationManifest{
		CreatedAt:     time.Now().UTC().Format(time.RFC3339),
		Provider:      providerValue,
		Model:         resolvedModel,
		StyleInput:    *style,
		StyleFile:     stylePath,
		Prompt:        *prompt,
		FinalPrompt:   finalPrompt,
		DryRun:        *dryRun,
		Character:     *character,
		CharacterFile: characterPath,
	}
	if providerValue == "openai" {
		manifest.Size = *size
		manifest.Quality = *quality
	}

	imagePath := filepath.Join(*outDir, "image-"+ts+".png")
	if !*dryRun {
		if providerValue == "google" {
			if err := ensureGoogleAPIKey(stdout, stderr); err != nil {
				fmt.Fprintf(stderr, "failed to configure Google API key: %v\n", err)
				return 1
			}
		}

		imageBytes, err := generateImage(providerValue, resolvedModel, finalPrompt, *size, *quality)
		if err != nil {
			fmt.Fprintf(stderr, "image generation failed: %v\n", err)
			return 1
		}

		if err := os.WriteFile(imagePath, imageBytes, 0o644); err != nil {
			fmt.Fprintf(stderr, "failed to write image: %v\n", err)
			return 1
		}

		manifest.ImagePath = imagePath
	}

	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		fmt.Fprintf(stderr, "failed to encode manifest: %v\n", err)
		return 1
	}

	manifestPath := filepath.Join(*outDir, "manifest-"+ts+".json")
	if err := os.WriteFile(manifestPath, data, 0o644); err != nil {
		fmt.Fprintf(stderr, "failed to write metadata: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Prompt: %s\n", finalPrompt)
	if *dryRun {
		fmt.Fprintln(stdout, "Dry run: image generation skipped.")
	} else {
		fmt.Fprintf(stdout, "Image saved: %s\n", imagePath)
	}
	fmt.Fprintf(stdout, "Manifest saved: %s\n", manifestPath)
	return 0
}

func resolveModel(provider string, override string) (string, error) {
	if override != "" {
		return override, nil
	}

	switch provider {
	case "google":
		return "gemini-2.5-flash-image", nil
	case "openai":
		return "gpt-image-1", nil
	default:
		return "", fmt.Errorf("unsupported provider %q (expected google or openai)", provider)
	}
}

func generateImage(provider string, model string, prompt string, size string, quality string) ([]byte, error) {
	switch provider {
	case "google":
		client, err := newGoogleClient()
		if err != nil {
			return nil, err
		}
		return client.generateImage(model, prompt)
	case "openai":
		client, err := newOpenAIClient()
		if err != nil {
			return nil, err
		}
		return client.generateImage(model, prompt, size, quality)
	default:
		return nil, fmt.Errorf("unsupported provider %q (expected google or openai)", provider)
	}
}

func normalizeGenerateArgs(args []string) []string {
	known := map[string]struct{}{
		"style":     {},
		"character": {},
		"prompt":    {},
		"out-dir":   {},
		"provider":  {},
		"model":     {},
		"size":      {},
		"quality":   {},
		"dry-run":   {},
		"h":         {},
		"help":      {},
	}

	normalized := make([]string, 0, len(args)+2)
	for _, arg := range args {
		if strings.HasPrefix(arg, "--") || !strings.HasPrefix(arg, "-") || arg == "-" {
			normalized = append(normalized, arg)
			continue
		}

		trimmed := strings.TrimPrefix(arg, "-")
		name := trimmed
		if idx := strings.Index(name, "="); idx >= 0 {
			name = name[:idx]
		}

		if _, ok := known[name]; ok {
			normalized = append(normalized, arg)
			continue
		}

		if strings.Contains(trimmed, "=") {
			normalized = append(normalized, arg)
			continue
		}

		normalized = append(normalized, "--character", trimmed)
	}

	return normalized
}
