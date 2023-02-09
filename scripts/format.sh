#!/bin/bash

which goimportssort || go install github.com/AanZee/goimportssort

for entry in `find . -name "*.go" | grep -v '.pb.'`; do
    echo $entry
    if grep -q "DO NOT EDIT" "$entry"; then
      continue
    fi
    goimportssort -w -local github.com/bnb-chain/ $entry
done
