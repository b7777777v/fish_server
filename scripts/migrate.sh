#!/bin/bash

# Exit immediately if a command exits with a non-zero status.
set -e

# Get the root directory of the project
PROJECT_ROOT=$(git rev-parse --show-toplevel)

# Define paths
MIGRATOR_SOURCE="${PROJECT_ROOT}/cmd/migrator"
MIGRATOR_BINARY="${PROJECT_ROOT}/bin/migrator"

# Check if a command is provided
if [ -z "$1" ]; then
  echo "Usage: $0 <up|down|force|version> [args]"
  echo "  up: Apply all available migrations."
  echo "  down: Revert the last migration."
  echo "  force <version>: Force migration to a specific version."
  echo "  version: Show the current migration version."
  exit 1
fi

# Build the migrator binary
echo "Building migrator..."
(cd "${PROJECT_ROOT}" && go build -o "${MIGRATOR_BINARY}" "${MIGRATOR_SOURCE}")

# Run the migrator with all provided arguments
echo "Running migrator with command: '$*' ..."
"${MIGRATOR_BINARY}" "$@"

echo "Migration script finished."
