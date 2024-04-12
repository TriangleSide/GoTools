#!/usr/bin/env bash

set -euo pipefail

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
echo "Creating a tarball of $SCRIPT_DIR"

TARBALL_LOCATION=".tarballs"
mkdir -p "$TARBALL_LOCATION"

CURRENT_DATETIME=$(date -u +'%Y-%m-%d-%H-%M-%S-UTC')
TARBALL_FILE="$TARBALL_LOCATION/intelligence_repo.$CURRENT_DATETIME.tar.gz"
ENCRYPTED_FILE="$TARBALL_LOCATION/intelligence-repo.$CURRENT_DATETIME.tar.gz.gpg"
PASSWORD_FILE=".enc_key.txt"

echo "Creating tarball into $TARBALL_FILE."
tar --exclude="$TARBALL_LOCATION" -czf "$TARBALL_FILE" .

echo "Encrypting into $ENCRYPTED_FILE"
gpg --batch --yes --passphrase-file "$PASSWORD_FILE" --symmetric --output "$ENCRYPTED_FILE" "$TARBALL_FILE"
rm "$TARBALL_FILE"

# To decrypt, run:
# gpg --batch --yes --passphrase-file "$PASSWORD_FILE" --output "intelligence-repo.tar.gz" --decrypt "$ENCRYPTED_FILE"

echo "Done."
