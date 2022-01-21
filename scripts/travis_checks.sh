#!/bin/bash

go test
golangci-lint run -E gosec  -E revive -E goconst --tests=false
