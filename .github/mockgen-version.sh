#!/usr/bin/env bash

# Lint version should match what is in .github/workflows/build-and-test.yml
MOCKGEN_VERSION=v1.6.0

# If command not found then install required version
mockgen >/dev/null 2>&1 || go install github.com/golang/mock/mockgen@$MOCKGEN_VERSION

# Show the full version line
mockgen -version

# Check version
found_version=$(mockgen -version | head -n 1 | awk '{print $1}')
[ $found_version == $MOCKGEN_VERSION ] || echo -e "\033[33mWarning!\033[0m mockgen version $found_version does not match version $MOCKGEN_VERSION used in github actions"
