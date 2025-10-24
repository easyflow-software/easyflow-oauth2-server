#!/bin/bash
# This script installs the necessary dependencies for the project.

set -e

# Color for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if go is installed
if ! command -v go &> /dev/null
then
    echo -e "${RED}Go could not be found. Please install Go and try again.${NC}"
    exit 1
fi

# Check if GOPATH is included in PATH
if [[ ":$PATH:" != *":$(go env GOPATH)/bin:"* ]]; then
    echo -e "${RED}GOPATH/bin is not in your PATH. Please add $(go env GOPATH)/bin to your PATH and try again.${NC}"
    exit 1
fi

# Install reflex
echo -e "${BLUE}Installing reflex...${NC}"
go install github.com/cespare/reflex@latest

# Install migrate
echo -e "${BLUE}Installing migrate...${NC}"
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Install sqlc
echo -e "${BLUE}Installing sqlc...${NC}"
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Install mockery
echo -e "${BLUE}Installing mockery...${NC}"
go install github.com/vektra/mockery/v3@v3.5.5

# Install swag
echo -e "${BLUE}Installing swag...${NC}"
go install github.com/swaggo/swag/cmd/swag@latest

# Install golangci-lint
echo -e "${BLUE}Installing golangci-lint...${NC}"
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.5.0

# Install pre-commit
echo -e "${BLUE}Installing pre-commit...${NC}"
if command -v pipx &> /dev/null; then
    pipx install pre-commit
else
    echo -e "${RED}pipx could not be found. Please install pipx and add this:\n\n  export PATH=\$PATH:/home/$(whoami)/.local/bin\n\nto your shell rc and try again.${NC}"
    exit 1
fi

# Install pre-commit hooks
echo -e "${BLUE}Installing pre-commit hooks...${NC}"
pre-commit install

echo -e "${GREEN}All dependencies installed successfully.${NC}"
