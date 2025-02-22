@echo off

echo Building...
set GOOS=linux
go build -o widgets
set GOOS=windows

echo Done.
