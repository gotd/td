#!/usr/bin/env bash

set -e
touch coverage.out

for d in $(go list ./... | grep -v vendor); do
    go test -coverprofile=profile.out -covermode=atomic "$d"
    if [[ -f profile.out ]]; then
        cat profile.out >> coverage.out
        rm profile.out
    fi
done
