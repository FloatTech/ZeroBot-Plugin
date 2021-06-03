go version
mips-linux-musl-gcc -v
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=auto
go mod tidy
GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=1 CC=mips-linux-musl-gcc CXX=mips-linux-musl-g++ go build -ldflags "-s -w" -o zerobot