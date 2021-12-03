// Package convert 转换txt到pb
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/RomiChan/protobuf/proto"

	"github.com/FloatTech/ZeroBot-Plugin/plugin_diana/data"
)

var (
	compo data.Composition
)

func init() {
	compo.Array = make([]string, 0, 64)
}

// 参数：txt文件位置 pb文件位置
func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		compo.Array = append(compo.Array, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		panic(err)
	}

	data, _ := proto.Marshal(&compo)
	f, err1 := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err1 == nil {
		defer f.Close()
		_, err2 := f.Write(data)
		if err2 == nil {
			fmt.Println("成功")
		} else {
			panic(err2)
		}
	} else {
		panic(err1)
	}
}
