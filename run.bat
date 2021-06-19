go version
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=auto
go mod tidy
::go build -ldflags="-s -w" -o ZeroBot-Plugin.exe
go run main.go
pause
