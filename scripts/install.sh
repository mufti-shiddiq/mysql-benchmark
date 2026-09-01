#!/usr/bin/env sh
set -eu

REPO="${MYSQL_BENCHMARK_REPO:-mufti-shiddiq/mysql-benchmark}"
VERSION="${MYSQL_BENCHMARK_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
BINARY_NAME="mysql-benchmark"

detect_os() {
	case "$(uname -s)" in
		Linux) echo "linux" ;;
		Darwin) echo "darwin" ;;
		*) echo "unsupported" ;;
	esac
}

detect_arch() {
	case "$(uname -m)" in
		x86_64|amd64) echo "amd64" ;;
		arm64|aarch64) echo "arm64" ;;
		*) echo "unsupported" ;;
	esac
}

need() {
	if ! command -v "$1" >/dev/null 2>&1; then
		echo "missing required command: $1" >&2
		exit 1
	fi
}

OS="$(detect_os)"
ARCH="$(detect_arch)"
if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
	echo "unsupported platform: $(uname -s) $(uname -m)" >&2
	exit 1
fi

need "tar"
if command -v curl >/dev/null 2>&1; then
	DOWNLOAD="curl -fsSL"
elif command -v wget >/dev/null 2>&1; then
	DOWNLOAD="wget -qO-"
else
	echo "missing required command: curl or wget" >&2
	exit 1
fi

if [ "$VERSION" = "latest" ]; then
	URL="https://github.com/${REPO}/releases/latest/download/mysql-benchmark-${OS}-${ARCH}.tar.gz"
else
	URL="https://github.com/${REPO}/releases/download/${VERSION}/mysql-benchmark-${OS}-${ARCH}.tar.gz"
fi

TMP_DIR="$(mktemp -d)"
cleanup() {
	rm -rf "$TMP_DIR"
}
trap cleanup EXIT

echo "Downloading ${URL}"
# shellcheck disable=SC2086
$DOWNLOAD "$URL" | tar -xz -C "$TMP_DIR"

if [ ! -f "$TMP_DIR/$BINARY_NAME" ]; then
	echo "release archive did not contain $BINARY_NAME" >&2
	exit 1
fi

mkdir -p "$INSTALL_DIR"
if [ -w "$INSTALL_DIR" ]; then
	mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
	chmod +x "$INSTALL_DIR/$BINARY_NAME"
else
	echo "Installing to $INSTALL_DIR requires sudo"
	sudo mv "$TMP_DIR/$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
	sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"
fi

echo "Installed $BINARY_NAME to $INSTALL_DIR/$BINARY_NAME"
"$INSTALL_DIR/$BINARY_NAME" --version
