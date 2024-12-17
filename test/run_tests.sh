#!/bin/bash

# Check if MOTHERDUCK_TOKEN is set
if [ -z "$MOTHERDUCK_TOKEN" ]; then
    echo "Error: MOTHERDUCK_TOKEN environment variable is not set"
    echo "Please set it first:"
    echo "export MOTHERDUCK_TOKEN=your-token-here"
    exit 1
fi

# Run the tests
cd "$(dirname "$0")"
go test -v -timeout 30m
