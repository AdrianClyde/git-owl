# git-owl

## Rules

- Always run commands inside the flox environment: `flox activate -- bash -c 'go build ./... && go test ./...'`. Do not run `go` directly outside of flox.
- Do not tell the user something is done until both build and tests pass.
- If the build or tests fail, fix the issue and re-run before reporting completion.
