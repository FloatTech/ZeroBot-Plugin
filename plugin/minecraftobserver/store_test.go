package minecraftobserver

import (
	"errors"
	"github.com/jinzhu/gorm"
	"testing"
)

func cleanTestData(t *testing.T) {
	err := dbInstance.sdb.Delete(&ServerStatus{}).Where("id > 0").Error
	if err != nil {
		t.Fatalf("cleanTestData() error = %v", err)
	}
	err = dbInstance.sdb.Delete(&ServerSubscribe{}).Where("id > 0").Error
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
		newSS1 := &ServerStatus{
			ServerAddr:  "dx.zhaomc.net",
			Description: "测试服务器",
			Players:     "1/20",
			Version:     "1.16.5",
			FaviconMD5:  "1234567",
		}
		newSS2 := &ServerStatus{
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

	})
	t.Run("update", func(t *testing.T) {
		cleanTestData(t)
		newSS := &ServerStatus{
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
		err = dbInstance.updateServerStatus(&ServerStatus{
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
	t.Run("delete", func(t *testing.T) {
		cleanTestData(t)
		newSS := &ServerStatus{
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
}
