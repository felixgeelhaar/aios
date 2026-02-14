#!/usr/bin/env bash
set -euo pipefail

echo "Running Roady helper smoke tests..."
bash ci/test_roady_helpers.sh

echo "Checking Roady project health..."
roady status
roady drift detect
roady debt summary

echo "Roady preflight passed."
