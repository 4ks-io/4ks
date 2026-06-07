#!/bin/bash
set -euo pipefail

url=https://local.4ks.io/api/_dev

typesense_health="$(curl -fsS http://127.0.0.1:8108/health)"
printf '%s\n' "$typesense_health"
test "$typesense_health" = '{"ok":true}'

curl -fsS -k -o /dev/null -X 'POST' "$url/init-search-collections" \
  -H 'accept: application/json' \
  -d ''

go run dev/data/upload.go -u $url/recipes -t foo -f dev/data/tiny_dataset.csv
