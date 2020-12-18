#!/usr/bin/env bash
linterVersion=v1.33.0
echo "INSTALLING: Linter"
echo ""
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $linterVersion