---
name: commit
description: Create a conventional commit for staged/unstaged changes
---

Create a git commit following the Conventional Commits specification. This project uses `go-semantic-release` which determines version bumps from commit messages.

## Format

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

## Types

| Type | When to use | Version bump |
|------|------------|--------------|
| `feat` | New feature or capability | Minor |
| `fix` | Bug fix | Patch |
| `docs` | Documentation only | None |
| `style` | Formatting, whitespace, etc. | None |
| `refactor` | Code change that neither fixes a bug nor adds a feature | None |
| `test` | Adding or updating tests | None |
| `chore` | Build process, CI, tooling, dependencies | None |

## Breaking changes

Append `!` after the type/scope to trigger a **major** version bump:

```
feat!: redesign diff viewer layout
```

Or include `BREAKING CHANGE:` in the footer:

```
feat: redesign diff viewer layout

BREAKING CHANGE: removed --compact flag
```

## Rules

1. Use lowercase for the type and description
2. Do not end the description with a period
3. Keep the description under 72 characters
4. Use the body to explain **why**, not what
5. Only use `feat` or `fix` when a release should be triggered — use `chore`, `docs`, `refactor`, etc. for non-release changes
6. Always run `go build ./...` and `go test ./...` before committing
7. Do not add a `Co-Authored-By` trailer to commit messages

## Steps

1. Run `go build ./...` and `go test ./...` to verify the build
2. Review staged and unstaged changes with `git status` and `git diff`
3. Stage the relevant files
4. Draft a conventional commit message based on the changes
5. Create the commit

$ARGUMENTS
