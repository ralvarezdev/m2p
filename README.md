# m2p

A Markdown-to-PDF command line converter in Go. It renders Markdown with a custom typographic style (Tokyo Night theme, embedded Inter and JetBrains Mono fonts, syntax-highlighted code) to PDF, HTML or both.

Module `github.com/ralvarezdev/m2p` (Go 1.26.3). Built with cobra, viper, goldmark + chroma, chromedp and go-pdf/fpdf.

## Features

- Convert one `.md` file or every `.md` file in a directory.
- Output `pdf` (default), `html` or `both`; paper `a4` (default), `letter`, `a3` or `legal`.
- Two PDF engines: `chromium` (headless Chrome via chromedp), `native` (pure Go via fpdf) or `auto` (default: Chromium, falling back to native if no browser is found).
- Optional page breaks before `h2` or `h3`, optional brand footer suppression, optional open of the result after conversion.
- Configuration from flags, `M2P_`-prefixed environment variables and a TOML file.

## Installation

```bash
go install github.com/ralvarezdev/m2p/cmd/m2p@latest
# or from a clone
task build        # needs https://taskfile.dev; or: go build -o dist/m2p ./cmd/m2p
```

GoReleaser builds Linux, Windows and macOS (amd64, arm64) releases on `v*` tags. The Chromium engine needs Chrome or Chromium installed; the native engine does not.

## Usage

```bash
m2p convert README.md                          # -> README.pdf
m2p convert README.md -o out.pdf
m2p convert README.md --format both --paper letter
m2p convert ./docs --engine native --page-break h2   # batch convert a folder
m2p convert README.md --open --no-footer
m2p config show -o json                        # resolved configuration and sources
m2p version
```

`convert` flags: `-o, --output` (default `<input>.pdf`), `-f, --format` (`pdf`, `html`, `both`), `--paper`, `--engine` (`auto`, `chromium`, `native`), `--page-break` (`none`, `h2`, `h3`), `--no-footer`, `--open`. Global flags: `--config <path>` and `-q, --quiet`.

## Configuration

The config file defaults to `~/.config/m2p/config.toml` (override with `--config`). Environment variables use the `M2P_` prefix, for example `M2P_PAPER`. Keys mirror the flags: `format`, `paper`, `engine`, `no-footer`, `open`, `page-break`, `quiet`.

## Project structure

```
cmd/m2p/             entry point
internal/cli/        cobra commands: root, convert, config, version, open
internal/converter/  Markdown parsing, template, highlighting, chromium and native renderers
internal/assets/     embedded fonts, CSS and HTML template
Taskfile.yml         dev tasks (build, install, test, lint, vet, snapshot)
.goreleaser.yml      release build config
.github/workflows/   ci.yml (build, vet, race tests on Linux/macOS/Windows, lint), release.yml
```

No test files were present in the snapshot checked.

## License

No license file is present in this repository.
