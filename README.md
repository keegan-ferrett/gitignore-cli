# Git Ignore

There are probably 100+ clis that do the same, but this is mine.

A simple tool to write a `.gitignore` file from https://github.com/github/gitignore.

## Installation

Install the latest release with curl (macOS and Linux, amd64 and arm64):

```
curl -fsSL https://raw.githubusercontent.com/keegan-ferrett/gitignore-cli/main/install.sh | sh
```

The script downloads the matching binary from the latest GitHub release and
installs it to `/usr/local/bin`. To customise:

- `VERSION=v0.1.0` — pin to a specific release tag.
- `INSTALL_DIR=$HOME/.local/bin` — install somewhere other than `/usr/local/bin`.

```
curl -fsSL https://raw.githubusercontent.com/keegan-ferrett/gitignore-cli/main/install.sh | INSTALL_DIR=$HOME/.local/bin sh
```

## Usage

List the templates avaliable from https://github.com/github/gitignore.

```
gi list
```

Fetch and write a `.gitignore` file from https://github.com/github/gitignore.

```
gi C++
```
