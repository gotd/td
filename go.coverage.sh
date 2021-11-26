#!/usr/bin/env bash

set -e

go test -v -coverpkg=./... -coverprofile=profile_all.out ./...

# Filter most generated code.
# Reduces size from 864M to 29M.
grep -v -P 'tl_.+_gen\.go' profile_all.out > profile.out
rm profile_all.out
