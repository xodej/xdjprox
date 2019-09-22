cls
go env

set GOARCH=amd64
set GOOS=linux
go build -ldflags="-s -w"
PAUSE