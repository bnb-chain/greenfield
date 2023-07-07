#!/usr/bin/env bash

mockgencmd="mockgen"

${mockgencmd} -source=x/challenge/types/expected_keepers.go -destination=x/challenge/types/expected_keepers_mocks.go -package=types
${mockgencmd} -source=x/payment/types/expected_keepers.go -destination=x/payment/types/expected_keepers_mocks.go -package=types
${mockgencmd} -source=x/permission/types/expected_keepers.go -destination=x/permission/types/expected_keepers_mocks.go -package=types
${mockgencmd} -source=x/sp/types/expected_keepers.go -destination=x/sp/types/expected_keepers_mocks.go -package=types
${mockgencmd} -source=x/storage/types/expected_keepers.go -destination=x/storage/types/expected_keepers_mocks.go -package=types
${mockgencmd} -source=x/virtualgroup/types/expected_keepers.go -destination=x/virtualgroup/types/expected_keepers_mocks.go -package=types
