go build -ldflags "-s -w" -o rename

GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o rename

GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o rename