#!/bin/bash
#MISE description="Build the CLI locally (pass `--snapshot` to build HEAD)"

goreleaser build --clean "$@"
