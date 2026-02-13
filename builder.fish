#!/usr/bin/env fish
cd builder
go build -o ../phis-builder main.go
cd ..
./phis-builder $argv
