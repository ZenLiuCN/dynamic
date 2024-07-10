#!/bin/sh
go list -export -f '{{if .Export}}packagefile {{.ImportPath}}={{.Export}}{{end}}' \
std `go list -f {{.Imports}} $1 | awk '{sub(/^\[/, ""); print }' | awk '{sub(/\]$/, ""); print }'` \
> importcfg
go tool compile -importcfg importcfg $1
rm -rf importcfg
