#!/bin/sh

set -e # Fail and exit on error of any command

REPO_OWNER='thetillhoff'
REPO_NAME='temingo'
CLI_NAME='temingo'

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

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
  printf "Either curl or wget are required to run this script"
  exit 2
fi

LATEST_VERSION="$($DOWNLOAD_BODY_CMD https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest | jq -r '.tag_name')"
echo https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_VERSION}/${CLI_NAME}_${OS}_${ARCH}
echo https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_VERSION}/${CLI_NAME}_${OS}_${ARCH}.sha256
$DOWNLOAD_FILE_CMD ${CLI_NAME} "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_VERSION}/${CLI_NAME}_${OS}_${ARCH}"
$DOWNLOAD_FILE_CMD ${CLI_NAME}.sha256 "https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${LATEST_VERSION}/${CLI_NAME}_${OS}_${ARCH}.sha256"
printf "$(cat ${CLI_NAME}.sha256) ${CLI_NAME}" | sha256sum --check --status
printf "Checksum validation complete, installing to /usr/local/bin/ ...\n"
sudo install ${CLI_NAME} /usr/local/bin/${CLI_NAME} # automatically sets rwxr-xr-x permissions
rm ${CLI_NAME} ${CLI_NAME}.sha256