go version
gcc -v
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=auto
go build -ldflags="-s -w" -o ZeroBot-Plugin-windows.exe
pause
