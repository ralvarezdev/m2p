# m2p — agent guide

A Markdown-to-PDF command line converter in Go (`github.com/ralvarezdev/m2p`):
renders one `.md` file or a whole folder to PDF, HTML or both with a fixed Tokyo
Night typographic theme, embedded Inter/JetBrains Mono and syntax-highlighted
code. Single-user developer tool; no test files and no LICENSE file yet.

## Commands

- Build: `task build` → `dist/m2p`; `task install` → `$GOPATH/bin`
- Test / vet / lint: `task test`, `task vet`, `task lint`
- Try a conversion: `task test:convert INPUT=my.md ENGINE=native PAGE_BREAK=h2`
  (defaults to `testdata/sample.md` → `dist/sample.pdf`), or
  `task test:convert:dir INPUT=./docs`
- All-platform local build: `task snapshot` (goreleaser, local only)
- Release: `task tag VERSION=0.1.0` pushes the tag that starts `release.yml`

`.github/workflows/ci.yml` builds, vets, runs `go test -race ./...` on Linux, macOS
and Windows, then lints.

## Layout

```
cmd/m2p/             entry point
internal/cli/        cobra commands: root, convert, config, version, open
internal/converter/  markdown, template, highlighting, chromium and native renderers
internal/assets/     embedded fonts, styles.css, document.html
testdata/            sample.md (gitignored)
docs/concept-design.html
```

## Conventions

- Two PDF engines behind one path: `chromium` (chromedp) and `native` (fpdf).
  `auto` uses Chromium and silently falls back to native when no browser is found,
  so a layout bug can be invisible in the default run — check `--engine chromium`
  before blaming the native renderer.
- A new option needs three places: the cobra flag, the `M2P_`-prefixed env key and
  the TOML key in `~/.config/m2p/config.toml`. Precedence is flag > env > file.
- Data goes to stdout; progress and errors go to stderr (`infof` respects `--quiet`).
- The embedded CSS and template in `internal/assets` are the product: a visual
  tweak changes the output of every document.

## Boundaries

- Always: run `task lint vet test` before declaring done.
- Ask first: changing the theme or template; tagging a release (`task tag` pushes
  and GoReleaser publishes).
- Never: commit `dist/`, `testdata/`, `docs/internal/` or `.env`; make the native
  engine require Chromium.
