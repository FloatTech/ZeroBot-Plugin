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
		msg, iconBase64 := resp.genServerSubscribeSchema("cn.nekoland.top", 123456).generateServerStatusMsg()
		fmt.Printf("msg: %v\n", msg)
		fmt.Printf("iconBase64: %v\n", iconBase64)
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
