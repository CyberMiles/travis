#!/usr/bin/env bash
DEPLOY=false

if [[ "$TRAVIS_EVENT_TYPE" == "cron" && "$TRAVIS_BRANCH" == "develop" && ($(($(date +%V) % 2)) == 0) ]]; then
  DEPLOY=true
fi


