#!/bin/bash
MAIN_VERSION=v$(grep "version =" cmd/ain/main.go | cut -f 4 -d " " | tr -d "\"")
if [ "$MAIN_VERSION" != "$1" ]; then
  echo "Tag version and version in main.go are not the same"
  exit 1
fi
