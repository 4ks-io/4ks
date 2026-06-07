#!/bin/bash
# Only regenerate PACKAGE_JSON when pnpm-lock.yaml actually changed.
# Avoids mass inotify events (delete + recreate) that trigger VSCode re-indexing on every Tilt start.
STAMP=.PACKAGE_JSON.stamp
LOCKFILE=pnpm-lock.yaml

if [ -f "$STAMP" ] && [ -f "$LOCKFILE" ] && diff -q "$STAMP" "$LOCKFILE" > /dev/null 2>&1; then
  exit 0
fi

rm -rf PACKAGE_JSON
find . -type d \( -name node_modules -o -name PACKAGE_JSON \) -prune -false -o -name package.json -exec bash -c 'path={}; d=./PACKAGE_JSON/$(dirname $path); mkdir -p $d ; cp $path $d' \;
find . -type d \( -name node_modules -o -name PACKAGE_JSON \) -prune -false -o -name project.json -exec bash -c 'path={}; d=./PACKAGE_JSON/$(dirname $path); mkdir -p $d ; cp $path $d' \;

cp "$LOCKFILE" "$STAMP"
