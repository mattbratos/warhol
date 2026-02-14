package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type styleProfile struct {
	Name           string   `yaml:"name"`
	Description    string   `yaml:"description"`
	PromptPrefix   []string `yaml:"prompt_prefix"`
	NegativePrompt []string `yaml:"negative_prompt"`
}

type characterProfile struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Traits      []string `yaml:"traits"`
	Outfit      []string `yaml:"outfit"`
	Prompt      string   `yaml:"prompt"`
}

func loadStyleProfile(nameOrPath string) (styleProfile, string, error) {
	path, err := resolveProfilePath("styles", nameOrPath)
	if err != nil {
		return styleProfile{}, "", err
	}

	var profile styleProfile
	if err := loadYAML(path, &profile); err != nil {
		return styleProfile{}, "", err
	}

	if profile.Name == "" {
		profile.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	return profile, path, nil
}

func loadCharacterProfile(nameOrPath string) (characterProfile, string, error) {
	path, err := resolveProfilePath("characters", nameOrPath)
	if err != nil {
		return characterProfile{}, "", err
	}

	var profile characterProfile
	if err := loadYAML(path, &profile); err != nil {
		return characterProfile{}, "", err
	}

	if profile.Name == "" {
		profile.Name = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	}

	return profile, path, nil
}

func resolveProfilePath(defaultDir string, nameOrPath string) (string, error) {
	added := make(map[string]struct{}, 8)
	candidates := make([]string, 0, 8)
	add := func(path string) {
		if _, exists := added[path]; exists {
			return
		}
		added[path] = struct{}{}
		candidates = append(candidates, path)
	}

	add(nameOrPath)
	add(filepath.Join("..", nameOrPath))

	if filepath.Ext(nameOrPath) == "" {
		for _, root := range []string{".", ".."} {
			add(filepath.Join(root, defaultDir, nameOrPath+".yaml"))
			add(filepath.Join(root, defaultDir, nameOrPath+".yml"))
		}
	} else if !strings.Contains(nameOrPath, string(os.PathSeparator)) {
		for _, root := range []string{".", ".."} {
			add(filepath.Join(root, defaultDir, nameOrPath))
		}
	}

	for _, candidate := range candidates {
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", err
		}
	}

	return "", fmt.Errorf("profile not found: %s", nameOrPath)
}

func loadYAML(path string, out any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	if err := yaml.Unmarshal(data, out); err != nil {
		return fmt.Errorf("parse %s: %w", path, err)
	}

	return nil
}

func buildFinalPrompt(style styleProfile, character *characterProfile, prompt string) string {
	parts := make([]string, 0, 12)

	if style.Description != "" {
		parts = append(parts, style.Description)
	}
	parts = append(parts, style.PromptPrefix...)

	if character != nil {
		if character.Prompt != "" {
			parts = append(parts, character.Prompt)
		} else {
			if character.Description != "" {
				parts = append(parts, character.Description)
			}
			if len(character.Traits) > 0 {
				parts = append(parts, "Traits: "+strings.Join(character.Traits, ", "))
			}
			if len(character.Outfit) > 0 {
				parts = append(parts, "Outfit: "+strings.Join(character.Outfit, ", "))
			}
		}
	}

	parts = append(parts, prompt)

	if len(style.NegativePrompt) > 0 {
		parts = append(parts, "Avoid: "+strings.Join(style.NegativePrompt, ", "))
	}

	return strings.Join(filterNonEmpty(parts), ". ")
}

func filterNonEmpty(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed != "" {
			filtered = append(filtered, trimmed)
		}
	}
	return filtered
}
