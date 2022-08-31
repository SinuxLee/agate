#!/usr/bin/env bash
set -e

swag init -g api/rest/handler.go -d ./internal/ -o ./internal/api/rest/docs --parseInternal  --generatedTime
