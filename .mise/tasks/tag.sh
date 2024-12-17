#!/bin/bash
#MISE description="Tag a new version"
#MISE depends=["build"]

git tag "$(svu next)" && git push --tags "$@"
