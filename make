#!/usr/bin/env bash

package=main

export GOPATH=$(pwd)

export CGO_ENABLED=0

export GOOS=darwin
export GOARCH=386
go install $package

export GOOS=linux
export GOARCH=386
go install $package

export GOOS=windows
export GOARCH=386
go install $package

export GOOS=darwin
export GOARCH=amd64
go install $package

export GOOS=linux
export GOARCH=amd64
go install $package

export GOOS=windows
export GOARCH=amd64
go install $package
