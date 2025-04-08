package minecraftobserver

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
	"testing"
)

func cleanTestData(t *testing.T) {
	err := dbInstance.sdb.Delete(&serverStatus{}).Where("id > 0").Error
	if err != nil {
		t.Fatalf("cleanTestData() error = %v", err)
	}
	err = dbInstance.sdb.Delete(&serverSubscribe{}).Where("id > 0").Error
	if err != nil {
		t.Fatalf("cleanTestData() error = %v", err)
	}
}

func Test_DAO(t *testing.T) {
	initErr := initializeDB("data/minecraftobserver/" + dbPath)
	if initErr != nil {
		t.Fatalf("initializeDB() error = %v", initErr)
	}
	if dbInstance == nil {
		t.Fatalf("initializeDB() got = %v, want not nil", dbInstance)
	}
	t.Run("insert", func(t *testing.T) {
		cleanTestData(t)
		newSS1 := &serverStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		newSS2 := &serverStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.8",
			FaviconMD5:  "1234567",
		}
		err := dbInstance.updateServerStatus(newSS1)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		err = dbInstance.updateServerStatus(newSS2)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}

		// check insert
		queryResult, err := dbInstance.getServerStatus("dx.zhaomc.net")
		if err != nil {
			t.Fatalf("getServerStatus() error = %v", err)
		}
		if queryResult == nil {
			t.Fatalf("getServerStatus() got = %v, want not nil", queryResult)
		}
		if queryResult.Version != "1.16.8" {
			t.Fatalf("getServerStatus() got = %v, want 1.16.8", queryResult.Version)
		}

		err = dbInstance.newSubscribe("dx.zhaomc.net", 123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		err = dbInstance.newSubscribe("dx.zhaomc.net", 123456, targetTypeUser)
		if err != nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		// check insert
		res, err := dbInstance.getAllSubscribes()
		if err != nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		if len(res) != 2 {
			t.Fatalf("getAllServer() got = %v, want 2", len(res))
		}
		// 检查是否符合预期
		if res[0].ServerAddr != "dx.zhaomc.net" {
			t.Fatalf("getAllServer() got = %v, want dx.zhaomc.net", res[0].ServerAddr)
		}
		if res[0].TargetType != targetTypeGroup {
			t.Fatalf("getAllServer() got = %v, want %v", res[0].TargetType, targetTypeGroup)
		}
		if res[1].ServerAddr != "dx.zhaomc.net" {
			t.Fatalf("getAllServer() got = %v, want dx.zhaomc.net", res[1].ServerAddr)
		}
		if res[1].TargetType != targetTypeUser {
			t.Fatalf("getAllServer() got = %v, want %v", res[1].TargetType, targetTypeUser)
		}

		// 顺带验证一下 byTarget
		res2, err := dbInstance.getSubscribesByTarget(123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("getSubscribesByTarget() error = %v", err)
		}
		if len(res2) != 1 {
			t.Fatalf("getSubscribesByTarget() got = %v, want 1", len(res2))
		}

	})
	// 重复添加订阅
	t.Run("insert dup", func(t *testing.T) {
		cleanTestData(t)
		newSS := &serverStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		err := dbInstance.updateServerStatus(newSS)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		err = dbInstance.newSubscribe("dx.zhaomc.net", 123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		err = dbInstance.newSubscribe("dx.zhaomc.net", 123456, targetTypeGroup)
		if err == nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		fmt.Printf("insert dup error: %+v", err)
	})

	t.Run("update", func(t *testing.T) {
		cleanTestData(t)
		newSS := &serverStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		err := dbInstance.updateServerStatus(newSS)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		err = dbInstance.updateServerStatus(&serverStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "更新测试",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		})
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		// check update
		queryResult2, err := dbInstance.getServerStatus("dx.zhaomc.net")
		if err != nil {
			t.Errorf("getAllServer() error = %v", err)
		}
		if queryResult2.Description != "更新测试" {
			t.Errorf("getAllServer() got = %v, want 更新测试", queryResult2.Description)
		}
	})
	t.Run("delete status", func(t *testing.T) {
		cleanTestData(t)
		newSS := &serverStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		err := dbInstance.updateServerStatus(newSS)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		// check insert
		queryResult, err := dbInstance.getServerStatus("dx.zhaomc.net")
		if err != nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		if queryResult == nil {
			t.Fatalf("getAllServer() got = %v, want not nil", queryResult)
		}
		err = dbInstance.delServerStatus("dx.zhaomc.net")
		if err != nil {
			t.Fatalf("deleteServerStatus() error = %v", err)
		}
		// check delete
		_, err = dbInstance.getServerStatus("dx.zhaomc.net")
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("getAllServer() error = %v", err)
		}

	})

	// 删除订阅
	t.Run("delete subscribe", func(t *testing.T) {
		cleanTestData(t)
		newSS := &serverStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		err := dbInstance.updateServerStatus(newSS)
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		err = dbInstance.newSubscribe("dx.zhaomc.net", 123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		err = dbInstance.deleteSubscribe("dx.zhaomc.net", 123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("deleteSubscribe() error = %v", err)
		}
		// check delete
		_, err = dbInstance.getServerStatus("dx.zhaomc.net")
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("getAllServer() error = %v", err)
		}
	})

	// 重复删除订阅
	t.Run("delete subscribe dup", func(t *testing.T) {
		cleanTestData(t)
		err := dbInstance.updateServerStatus(&serverStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		})
		if err != nil {
			t.Errorf("upsertServerStatus() error = %v", err)
		}
		err = dbInstance.newSubscribe("dx.zhaomc.net", 123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("newSubscribe() error = %v", err)
		}

		err = dbInstance.newSubscribe("dx.zhaomc.net123", 123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("newSubscribe() error = %v", err)
		}
		err = dbInstance.updateServerStatus(&serverStatus{
			ServerAddr:  "dx.zhaomc.net123",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		})
		if err != nil {
			t.Fatalf("updateServerStatus() error = %v", err)
		}
		err = dbInstance.newSubscribe("dx.zhaomc.net4567", 123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("newSubscribe() error = %v", err)
		}
		err = dbInstance.updateServerStatus(&serverStatus{
			ServerAddr:  "dx.zhaomc.net4567",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		})
		if err != nil {
			t.Fatalf("updateServerStatus() error = %v", err)
		}

		// 检查是不是3个
		allSub, err := dbInstance.getAllSubscribes()
		if err != nil {
			t.Fatalf("getAllSubscribes() error = %v", err)
		}
		if len(allSub) != 3 {
			t.Fatalf("getAllSubscribes() got = %v, want 3", len(allSub))
		}
		err = dbInstance.deleteSubscribe("dx.zhaomc.net", 123456, targetTypeGroup)
		if err != nil {
			t.Fatalf("deleteSubscribe() error = %v", err)
		}
		err = dbInstance.deleteSubscribe("dx.zhaomc.net", 123456, targetTypeGroup)
		if err == nil {
			t.Fatalf("deleteSubscribe() error = %v", err)
		}
		fmt.Println("delete dup error: ", err)

		// 检查其他的没有被删
		allSub, err = dbInstance.getAllSubscribes()
		if err != nil {
			t.Fatalf("getAllSubscribes() error = %v", err)
		}
		// 检查是否符合预期
		if len(allSub) != 2 {
			t.Fatalf("getAllSubscribes() got = %v, want 2", len(allSub))
		}
		// 状态
		_, err = dbInstance.getServerStatus("dx.zhaomc.net")
		if !gorm.IsRecordNotFoundError(err) {
			t.Fatalf("getAllServer() error = %v", err)
		}
		status1, err := dbInstance.getServerStatus("dx.zhaomc.net123")
		if err != nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		status2, err := dbInstance.getServerStatus("dx.zhaomc.net4567")
		if err != nil {
			t.Fatalf("getAllServer() error = %v", err)
		}
		if status1 == nil || status2 == nil {
			t.Fatalf("getAllServer() want not nil")
		}

	})
}
