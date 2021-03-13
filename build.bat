SET CGO_ENABLED=1
go build -ldflags="-s -w  -extldflags '-static'" main.go
pause
