#!/bin/bash

go install github.com/AanZee/goimportssort

for entry in `find . -name "*.go"`; do
    echo $entry
    if grep -q "DO NOT EDIT" "$entry"; then
      echo "xxxxxxxx=================================="
      continue
    fi
    goimportssort -w -local github.com/bnb-chain/ $entry
done