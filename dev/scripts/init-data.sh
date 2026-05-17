#!/bin/bash

url=https://local.4ks.io/api/_dev

curl -s -k -o /dev/null -X 'POST' "$url/init-search-collections" \
  -H 'accept: application/json' \
  -d ''

go run dev/data/upload.go -u $url/recipes -t foo -f dev/data/tiny_dataset.csv