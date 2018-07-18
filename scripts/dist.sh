#!/bin/bash
set -e

# this script is used in a docker container, don't run it directly.

apk update && apk add zip
zip "/scripts/travis_${BUILD_TAG}_linux_amd64.zip" /app/travis
