# Go setup for RobotFleetOS

Go is installed from the official tarball (no Homebrew required).

## Location

- **Go 1.22.4** is installed at: `~/go-sdk-1.22.4/`
- **Binary**: `~/go-sdk-1.22.4/bin/go`

The tarball `go1.22.4.darwin-amd64.tar.gz` was moved to your home directory (`~`).

## Use in this project

From the project root:

```bash
# Add Go to PATH for this session
export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"

# Optional: use project-local GOPATH so cache lives in the repo
export GOPATH="$(pwd)/.gopath"

go mod tidy
go build -o bin/fleet ./cmd/fleet
./bin/fleet
```

## Make it permanent (optional)

Add to your `~/.zshrc`:

```bash
export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"
```

Then open a new terminal or run `source ~/.zshrc`.

## Reinstall or install elsewhere

1. Download: https://go.dev/dl/go1.22.4.darwin-amd64.tar.gz (or newer from https://go.dev/dl/)
2. Extract: `tar -C "$HOME" -xzf go1.22.4.darwin-amd64.tar.gz`
3. Rename so it doesn’t conflict with a `~/go` workspace: `mv "$HOME/go" "$HOME/go-sdk-1.22.4"`
4. Use: `export PATH="$HOME/go-sdk-1.22.4/bin:$PATH"`
