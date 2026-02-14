# warhol

`warhol` is a CLI for creating images in a certain, visually consistent style, plus a docs/landing site.

## Install

Homebrew:

```bash
brew tap mattbratos/tap
brew install warhol
```

Run without installing globally:

```bash
make cli-run
```

## Repo Layout

- `cli/`: Go CLI source code.
- `www/`: Fumadocs + Next.js site.
- `Makefile`: root task runner for both apps.

## Quick Start

CLI:

```bash
export GEMINI_API_KEY=your_api_key_here

make cli-run CLI_ARGS='generate --style 16bit -matt --prompt "full body portrait, city street at night"'
```

This loads:
- style: `styles/16bit.yaml`
- character: `characters/matt.yaml` (from `-matt`)
- provider: `google` (default)
- model: `gemini-2.5-flash-image` (default, often called "Nano Banana")

The CLI writes:
- generated image: `outputs/image-<timestamp>.png`
- metadata: `outputs/manifest-<timestamp>.json`

Website:

```bash
make www-dev
```
