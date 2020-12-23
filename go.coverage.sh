#!/usr/bin/env bash

set -e
touch coverage.out

echo 'mode: atomic' > coverage.out
go list ./... | grep -v /cmd | grep -v /vendor | xargs -n1 -I{} sh -c 'go test -covermode=atomic -coverprofile=profile.out -coverpkg $(go list ./... ) {} && tail -n +2 profile.out >> coverage.out || exit 255' && rm coverage.tmp
