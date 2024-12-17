#!/bin/bash
#MISE description="Tag a new version"

git tag "$(svu next)" && git push --tags "$@"
