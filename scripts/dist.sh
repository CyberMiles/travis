#!/bin/bash
set -e

# this script is used in a docker container, don't run it directly.

apt-get update && apt-get -y install zip
zip "/scripts/travis_${BUILD_TAG}_linux_amd64.zip" /app/travis
