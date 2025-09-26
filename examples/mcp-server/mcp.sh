#!/usr/bin/env bash

SCRIPT_PATH="$(dirname -- "${BASH_SOURCE[0]}")"

cd "${SCRIPT_PATH}" || exit
go run "${SCRIPT_PATH}/." mcp "$@"