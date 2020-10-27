@echo off
set CGO_ENABLED=0
set GOOS=windows
set GOARCH=386
go build -o lsp.exe