go version
mips-linux-musl-gcc -v
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GO111MODULE=auto
go mod tidy
export CCBIN=~/openwrt_with_lean_packages/staging_dir/toolchain-mips_24kc_gcc-8.4.0_musl/bin/mips-openwrt-linux-musl-gcc
export CXXBIN=~/openwrt_with_lean_packages/staging_dir/toolchain-mips_24kc_gcc-8.4.0_musl/bin/mips-openwrt-linux-musl-g++
GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=1 CC=${CCBIN} CXX=${CXXBIN} go build -ldflags "-s -w" -o zerobot