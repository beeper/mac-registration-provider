#!/bin/sh
go build -ldflags "-X main.Commit=$(git rev-parse HEAD)"
