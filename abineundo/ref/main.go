// Package main generate necessary files that needed for compiling
package main

import (
	"flag"
	"io"
	"os"
	"path"
)

func main() {
	root := flag.String("r", "", "project root dir")
	flag.Parse()

	fi, err := os.Open(path.Join(*root, "main.go"))
	if err != nil {
		panic(err)
	}
	fo, err := os.Create(path.Join(*root, "abineundo/ref/main/main.go"))
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(fo, fi)
	if err != nil {
		panic(err)
	}
	fi.Close()
	fo.Close()

	regf := path.Join(*root, "custom/register.go")
	tgtf := path.Join(*root, "abineundo/ref/custom/register.go")
	if _, err := os.Stat(regf); err != nil {
		if os.IsNotExist(err) {
			_ = os.WriteFile(tgtf, []byte("// Package custom ...\npackage custom\n"), 0644)
			return
		}
		panic(err)
	}

	fi, err = os.Open(regf)
	if err != nil {
		panic(err)
	}
	fo, err = os.Create(tgtf)
	if err != nil {
		panic(err)
	}
	_, err = io.Copy(fo, fi)
	if err != nil {
		panic(err)
	}
	fi.Close()
	fo.Close()
}
