<img width="100%" alt="git-owl-logo" src="https://github.com/user-attachments/assets/37313e3b-c2fa-4040-b4aa-8357b7062872" />

Git Owl is a pretty terminal git diff viewer with an animated owl that judges your code.

<div align="center">
  <video src="https://github.com/user-attachments/assets/a78988b9-0114-4ad5-ab9a-49db48e03823" width="100%" />
</div>

## Why another git diff viewer?

Because why not.

## The real reason

More and more my workflow has simplified to just interacting with Claude and giving it passive aggressive feedback until it implements things correctly. Having a quick file viewer running next to Claude lets me yell at it faster.

This is that viewer.

## What it does

- Shows your changed files with syntax-highlighted diffs
- Auto-refreshes every 2 seconds so you can watch Claude butcher your codebase in real time
- Has an animated owl in the corner that blinks at you
- Tokyo Night theme because we have taste

## Install

Grab the latest binary from [Releases](https://github.com/AdrianClyde/git-owl/releases).

Or build from source:

```bash
flox activate
go build -o git-owl
```

> This project uses [Flox](https://flox.dev) to manage dependencies. `flox activate` drops you into a shell with Go, Git, and everything else you need.

## Usage

```bash
# Run in current repo
git-owl

# Run against a specific repo
git-owl /path/to/repo
```

## Keybindings

| Key | Action |
|-----|--------|
| `j/k` or `↑/↓` | Navigate files |
| `Enter` | View file |
| `d` | Toggle diff view |
| `p` | Toggle markdown preview |
| `t` | Toggle all files / changed only |
| `/` | Filter files |
| `?` | Help |
| `q` | Quit |

## Built with

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling
- [Chroma](https://github.com/alecthomas/chroma) — Syntax highlighting
- [Glamour](https://github.com/charmbracelet/glamour) — Markdown rendering
- Mass quantities of passive aggression

## License

MIT
