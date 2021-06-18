go version
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=auto
go mod tidy
GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 CC=${CCBIN} CXX=${CXXBIN} go build -ldflags "-s -w" -o zerobot