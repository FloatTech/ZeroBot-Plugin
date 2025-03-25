package minecraftobserver

import (
	"fmt"
	"github.com/wdvxdr1123/ZeroBot/message"
	"testing"
)

func Test_singleServerScan(t *testing.T) {
	initErr := initializeDB("data/minecraftobserver/" + dbPath)
	if initErr != nil {
		t.Fatalf("initializeDB() error = %v", initErr)
	}
	if dbInstance == nil {
		t.Fatalf("initializeDB() got = %v, want not nil", dbInstance)
	}
	t.Run("状态变更", func(t *testing.T) {
		cleanTestData(t)
		newSS1 := &serverStatus{
			ServerAddr:  "cn.nekoland.top",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "",
		}
		err := dbInstance.updateServerStatus(newSS1)
		if err != nil {
			t.Fatalf("upsertServerStatus() error = %v", err)
		}
		err = dbInstance.newSubscribe("cn.nekoland.top", 123456, 1)
		if err != nil {
			t.Fatalf("getServerSubscribeByTargetGroupAndAddr() error = %v", err)
		}
		changed, msg, err := singleServerScan(newSS1)
		if err != nil {
			t.Fatalf("singleServerScan() error = %v", err)
		}
		if !changed {
			t.Fatalf("singleServerScan() got = %v, want true", changed)
		}
		if len(msg) == 0 {
			t.Fatalf("singleServerScan() got = %v, want not empty", msg)
		}
		fmt.Printf("msg: %v\n", msg)
	})

	t.Run("可达 -> 不可达", func(t *testing.T) {
		cleanTestData(t)
		newSS1 := &serverStatus{
			ServerAddr:  "dx.123213213123123.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "",
			PingDelay:   123,
		}
		err := dbInstance.updateServerStatus(newSS1)
		if err != nil {
			t.Fatalf("upsertServerStatus() error = %v", err)
		}
		err = dbInstance.newSubscribe("dx.123213213123123.net", 123456, 1)
		if err != nil {
			t.Fatalf("getServerSubscribeByTargetGroupAndAddr() error = %v", err)
		}
		var msg message.Message
		changed, _, err := singleServerScan(newSS1)
		if err != nil {
			t.Fatalf("singleServerScan() error = %v", err)
		}
		if changed {
			t.Fatalf("singleServerScan() got = %v, want false", changed)
		}
		// 第二次
		changed, _, err = singleServerScan(newSS1)
		if err != nil {
			t.Fatalf("singleServerScan() error = %v", err)
		}
		if changed {
			t.Fatalf("singleServerScan() got = %v, want false", changed)
		}
		// 第三次
		changed, msg, err = singleServerScan(newSS1)
		if err != nil {
			t.Fatalf("singleServerScan() error = %v", err)
		}
		if !changed {
			t.Fatalf("singleServerScan() got = %v, want true", changed)
		}
		if len(msg) == 0 {
			t.Fatalf("singleServerScan() got = %v, want not empty", msg)
		}
		fmt.Printf("msg: %v\n", msg)

	})

	t.Run("不可达 -> 可达", func(t *testing.T) {
		cleanTestData(t)
		newSS1 := &serverStatus{
			ServerAddr:  "cn.nekoland.top",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "",
			PingDelay:   pingDelayUnreachable,
		}
		err := dbInstance.updateServerStatus(newSS1)
		if err != nil {
			t.Fatalf("upsertServerStatus() error = %v", err)
		}
		err = dbInstance.newSubscribe("cn.nekoland.top", 123456, 1)
		if err != nil {
			t.Fatalf("newSubscribe() error = %v", err)
		}
		changed, msg, err := singleServerScan(newSS1)
		if err != nil {
			t.Fatalf("singleServerScan() error = %v", err)
		}
		if !changed {
			t.Fatalf("singleServerScan() got = %v, want true", changed)
		}
		if len(msg) == 0 {
			t.Fatalf("singleServerScan() got = %v, want not empty", msg)
		}
		fmt.Printf("msg: %v\n", msg)
	})

}
