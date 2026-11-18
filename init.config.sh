#!/usr/bin/env bash

PROJECT_NAME="wialon-sener-pilot"
REQUIRED_TOOLS=("go")
REQUIRED_ENV_VARS=()
INSTALL_CMD=""
BUILD_CMD="go build ./..."
TEST_CMD="go test -count=1 ./..."
LINT_CMD="go vet ./..."
TYPECHECK_CMD=""
