#!/bin/bash
#MISE description="Release a new version"
#MISE depends=["build"]

GITHUB_TOKEN=$(gh auth token) goreleaser release --auto-snapshot --clean
