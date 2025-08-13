#!/bin/bash
# Script to create database migrations

# Add Go bin to PATH
export PATH=$PATH:/home/alex/go/bin

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null
then
    echo "migrate tool is not installed. Please install it first:"
    echo "go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Check if migration name is provided
if [ $# -eq 0 ]
then
    echo "Usage: $0 <migration-name>"
    exit 1
fi

# Create migration
migrate create -ext sql -dir backend/migrations -seq "$1"