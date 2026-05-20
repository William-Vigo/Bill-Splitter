#!/bin/bash

set -e

BUMP=$1

if [[ -z "$BUMP" ]]; then
  echo "Usage: sh version.sh [major|minor|patch]"
  exit 1
fi

CURRENT=$(cat VERSION)
CURRENT=${CURRENT#v}

IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT"

case $BUMP in
  major)
    ((MAJOR+=1))
    MINOR=0
    PATCH=0
    ;;
  minor)
    ((MINOR+=1))
    PATCH=0
    ;;
  patch)
    ((PATCH+=1))
    ;;
  *)
    echo "Invalid bump type: $BUMP"
    exit 1
    ;;
esac

NEW_VERSION="v$MAJOR.$MINOR.$PATCH"

echo "$NEW_VERSION" > VERSION

git add VERSION
git commit -m "$NEW_VERSION"

echo "Created release $NEW_VERSION"