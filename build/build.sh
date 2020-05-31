#!/bin/sh

export TARGET=$1

if [ "$TARGET" != "rc" ] && [ "$TARGET" != "remote-control" ]; then
  echo "Invalid argument.  Must be 'rc' or 'remote-control'"
  exit 1
fi

# path to the directory this script is in (based on how it was called)
SCRIPT_DIR=$(dirname "$0")

# get the version
export VERSION=$(cat $SCRIPT_DIR/versions.json | jq -r ".\"$TARGET\"")

mage build
