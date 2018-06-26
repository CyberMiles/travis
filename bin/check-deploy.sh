#!/usr/bin/env bash

VERSION=$(git rev-parse --short HEAD)

DEPLOY=false

# biweekly, on Thursday
if [[ "$TRAVIS_EVENT_TYPE" == "cron" \
    && "$TRAVIS_BRANCH" == "develop" \
    && ($(($(date +%V) % 2)) == 1) \
    && ($(date +%w) == 4) ]]; then
  DEPLOY=true
fi

echo $VERSION
echo $DEPLOY
