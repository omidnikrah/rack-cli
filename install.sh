#!/usr/bin/env bash
#
# Installs the latest rack-cli release.
#
#   curl -fsSL https://raw.githubusercontent.com/omidnikrah/rack-cli/main/install.sh | bash
#
# Environment:
#   RACK_CLI_VERSION      release tag to install, e.g. v0.1.0 (default: latest)
#   RACK_CLI_INSTALL_DIR  where to put the binary (default: /usr/local/bin,
#                         or ~/.local/bin when that is not writable)

set -euo pipefail

REPO="omidnikrah/rack-cli"
BIN="rack-cli"

info() { printf '==> %s\n' "$*"; }
fail() { printf 'error: %s\n' "$*" >&2; exit 1; }

need() {
  command -v "$1" >/dev/null 2>&1 || fail "'$1' is required but not installed"
}

detect_platform() {
  case "$(uname -s)" in
    Darwin) os=darwin ;;
    Linux)  os=linux ;;
    *) fail "unsupported OS: $(uname -s) (macOS and Linux only)" ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64)  arch=amd64 ;;
    arm64|aarch64) arch=arm64 ;;
    *) fail "unsupported architecture: $(uname -m)" ;;
  esac
}

release_url() {
  if [ -n "${RACK_CLI_VERSION:-}" ]; then
    echo "https://github.com/${REPO}/releases/download/${RACK_CLI_VERSION}"
  else
    echo "https://github.com/${REPO}/releases/latest/download"
  fi
}

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d ' ' -f 1
  else
    shasum -a 256 "$1" | cut -d ' ' -f 1
  fi
}

verify_checksum() {
  local dir="$1" archive="$2" expected actual

  expected="$(grep " ${archive}\$" "${dir}/checksums.txt" | cut -d ' ' -f 1)"
  [ -n "$expected" ] || fail "no checksum listed for ${archive}"

  actual="$(sha256_of "${dir}/${archive}")"
  [ "$expected" = "$actual" ] || fail "checksum mismatch for ${archive}"
}

pick_install_dir() {
  if [ -n "${RACK_CLI_INSTALL_DIR:-}" ]; then
    echo "$RACK_CLI_INSTALL_DIR"
  elif [ -w /usr/local/bin ]; then
    echo /usr/local/bin
  else
    echo "${HOME}/.local/bin"
  fi
}

main() {
  need curl
  need tar
  detect_platform

  local archive="${BIN}_${os}_${arch}.tar.gz"
  local base tmp dest
  base="$(release_url)"
  tmp="$(mktemp -d)"
  trap "rm -rf '$tmp'" EXIT

  info "Downloading ${archive}"
  curl -fsSL "${base}/${archive}" -o "${tmp}/${archive}" \
    || fail "could not download ${base}/${archive}"
  curl -fsSL "${base}/checksums.txt" -o "${tmp}/checksums.txt" \
    || fail "could not download checksums.txt"

  info "Verifying checksum"
  verify_checksum "$tmp" "$archive"
  tar -xzf "${tmp}/${archive}" -C "$tmp" "$BIN"

  dest="$(pick_install_dir)"
  info "Installing to ${dest}"
  mkdir -p "$dest" 2>/dev/null || sudo mkdir -p "$dest"
  if [ -w "$dest" ]; then
    install -m 755 "${tmp}/${BIN}" "${dest}/${BIN}"
  else
    sudo install -m 755 "${tmp}/${BIN}" "${dest}/${BIN}"
  fi

  info "Installed ${dest}/${BIN}"
  case ":${PATH}:" in
    *":${dest}:"*) ;;
    *) info "Note: ${dest} is not on your PATH" ;;
  esac
}

main "$@"
