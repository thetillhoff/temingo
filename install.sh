#!/bin/sh

set -e # Fail and exit on error of any command

REPO_OWNER='thetillhoff'
REPO_NAME='temingo'
CLI_NAME='temingo'

OS="$(uname -s | tr '[:upper:]' '[:lower:]')" # f.e. 'darwin'
# Verify OS
case "${OS}" in
  darwin|linux|windows) ;;
  *) cat <<EOF
Unsupported OS type ${OS} detected. Supported are darwin, linux, windows.
Feel free to open an issue or PR for your OS at https://github.com/thetillhoff/${REPO_NAME}."
EOF
  exit 0 ;;
esac

ARCH="$(uname -m)" # f.e. 'arm64'
if [ "${ARCH}" = "x86_64" ]; then # Overwrite ARCH, required for WSL
  ARCH="amd64"
fi
# Verify ARCH
case "${ARCH}" in
  amd64|arm64) ;;
  *) cat <<EOF
Unsupported ARCH type ${ARCH} detected. Supported are amd64, arm64.
Feel free to open an issue or PR for your ARCH at https://github.com/thetillhoff/${REPO_NAME}.
EOF
  exit 0 ;;
esac


# Check if jq is available
if [ ! 'command -v jq' ]; then
  printf "jq is required to run this script"
  exit 1
fi

# Check if sha256sum is available
if [ ! 'command -v sha256sum' ]; then
  printf "sha256sum is required to run this script"
  exit 1
fi

# Check if curl or wget are available
if [ 'command -v curl' ]; then
  DOWNLOAD_FILE_CMD="curl -Lo"
  DOWNLOAD_BODY_CMD="curl -sL"
elif [ 'command -v wget' ]; then
  DOWNLOAD_FILE_CMD="wget -O"
  DOWNLOAD_BODY_CMD="wget -qO-"
else
  printf "Neither curl or wget are available, please install one of them to run this script"
  exit 2
fi

# Create temporary directory for downloads
TEMP_DIR="$(mktemp -d)"

# Always remove temporary directory on exit
trap 'rm -rf "$TEMP_DIR"' EXIT

LATEST_VERSION="$($DOWNLOAD_BODY_CMD https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest | jq -r '.tag_name')"
printf "Downloading ${CLI_NAME} ${LATEST_VERSION} for ${OS} ${ARCH}\n"
$DOWNLOAD_FILE_CMD "${TEMP_DIR}/${CLI_NAME}" "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_VERSION}/${CLI_NAME}_${OS}_${ARCH}"
$DOWNLOAD_FILE_CMD "${TEMP_DIR}/${CLI_NAME}.sha256" "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_VERSION}/${CLI_NAME}_${OS}_${ARCH}.sha256"
printf "%s  %s\n" "$(cat "${TEMP_DIR}/${CLI_NAME}.sha256")" "${CLI_NAME}" | (cd "$TEMP_DIR" && sha256sum --check -)
printf "Checksum validation complete, installing to /usr/local/bin/ ...\n"
sudo install "${TEMP_DIR}/${CLI_NAME}" /usr/local/bin/${CLI_NAME} # automatically sets rwxr-xr-x permissions
printf "Installation complete!\n"
