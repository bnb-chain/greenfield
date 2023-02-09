#!/bin/bash

go install github.com/AanZee/goimportssort

for entry in `find . -name "*.go"`; do
    if grep -q "DO NOT EDIT" "$entry"; then
      continue
    fi
    goimportssort -w -local github.com/bnb-chain/ $entry
done