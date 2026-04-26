#!/usr/bin/env sh
# Installer for the gitignore CLI.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/keegan-ferrett/gitignore-cli/main/install.sh | sh
#
# Environment overrides:
#   VERSION       Release tag to install (default: latest, e.g. "v0.1.0").
#   INSTALL_DIR   Target directory for the binary (default: /usr/local/bin).

set -eu

REPO="keegan-ferrett/gitignore-cli"
BINARY="gitignore"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-}"

err() { printf 'error: %s\n' "$*" >&2; exit 1; }
info() { printf '%s\n' "$*"; }

# detect_os maps `uname -s` to the GOOS values used in release artifacts.
detect_os() {
    case "$(uname -s)" in
        Darwin) echo "darwin" ;;
        Linux) echo "linux" ;;
        *) err "unsupported OS: $(uname -s) (only macOS and Linux are supported)" ;;
    esac
}

# detect_arch maps `uname -m` to the GOARCH values used in release artifacts.
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64) echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *) err "unsupported architecture: $(uname -m) (supported: x86_64, arm64)" ;;
    esac
}

# resolve_version returns VERSION if set, otherwise queries the GitHub
# releases API for the latest tag.
resolve_version() {
    if [ -n "$VERSION" ]; then
        echo "$VERSION"
        return
    fi
    api="https://api.github.com/repos/${REPO}/releases/latest"
    tag=$(curl -fsSL "$api" | sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1)
    [ -n "$tag" ] || err "could not determine latest release from $api"
    echo "$tag"
}

require() {
    command -v "$1" >/dev/null 2>&1 || err "$1 is required but not installed"
}

main() {
    require curl
    require tar

    os=$(detect_os)
    arch=$(detect_arch)
    version=$(resolve_version)

    archive="${BINARY}_${version}_${os}_${arch}.tar.gz"
    url="https://github.com/${REPO}/releases/download/${version}/${archive}"

    info "Installing ${BINARY} ${version} for ${os}/${arch}"
    info "  source: ${url}"
    info "  target: ${INSTALL_DIR}/${BINARY}"

    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT INT TERM

    curl -fsSL "$url" -o "${tmpdir}/${archive}" \
        || err "download failed; check that ${version} exists at https://github.com/${REPO}/releases"
    tar -xzf "${tmpdir}/${archive}" -C "$tmpdir"

    # Use sudo when the install directory is not writable by the current user.
    if [ -w "$INSTALL_DIR" ] || [ ! -e "$INSTALL_DIR" ]; then
        mkdir -p "$INSTALL_DIR"
        mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    else
        info "  ${INSTALL_DIR} is not writable; using sudo"
        sudo mkdir -p "$INSTALL_DIR"
        sudo mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
    fi

    info "Installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"

    # Warn if the install dir is not on PATH so the user knows why the binary
    # is "missing" right after install.
    case ":$PATH:" in
        *":${INSTALL_DIR}:"*) ;;
        *) info "Note: ${INSTALL_DIR} is not on your PATH. Add it to your shell rc file." ;;
    esac
}

main "$@"
