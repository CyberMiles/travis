#!/bin/bash
set -e

# this script is used in a docker container, don't run it directly.

apt-get update && apt-get -y install zip wget tar

cd /app
zip "/scripts/travis_${BUILD_TAG}_ubuntu-16.04.zip" travis* lib/*
