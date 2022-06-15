#!/usr/bin/env bash

set -eu
cmd=kubecon-cni
eval $(go env | grep -e "GOHOSTOS" -e "GOHOSTARCH")
GO=${GO:-go}
GOOS=${GOOS:-${GOHOSTOS}}
GOARCH=${GOARCH:-${GOHOSTARCH}}
GOFLAGS=${GOFLAGS:-}

CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} ${GO} build ${GOFLAGS} -o bin/${cmd} cmd/${cmd}.go
