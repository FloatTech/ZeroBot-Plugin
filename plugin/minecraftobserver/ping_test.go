package minecraftobserver

import (
	"fmt"
	"testing"
)

func Test_PingListInfo(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		resp, err := getMinecraftServerStatus("cn.nekoland.top")
		if err != nil {
			t.Fatalf("getMinecraftServerStatus() error = %v", err)
		}
		msg := resp.GenServerSubscribeSchema("cn.nekoland.top", 123456).GenerateServerStatusMsg()
		fmt.Printf("msg: %v\n", msg)
	})
	t.Run("不可达", func(t *testing.T) {
		ss, err := getMinecraftServerStatus("dx.123213213123123.net")
		if err == nil {
			t.Fatalf("getMinecraftServerStatus() error = %v", err)
		}
		if ss != nil {
			t.Fatalf("getMinecraftServerStatus() got = %v, want nil", ss)
		}
	})
}
