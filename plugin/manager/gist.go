package manager

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/wdvxdr1123/ZeroBot/utils/helper"

	"github.com/FloatTech/floatbox/math"
	"github.com/FloatTech/floatbox/web"
)

// user hash file
const gistraw = "https://gist.githubusercontent.com/%s/%s/raw/%s"

func checkNewUser(qq, gid int64, ghun, hash string) (bool, string) {
	if db.CanFind("member", "where ghun="+ghun) {
		return false, "该github用户已入群"
	}
	gidsum := md5.Sum(helper.StringToBytes(strconv.FormatInt(gid, 10)))
	gidhex := hex.EncodeToString(gidsum[:])
	u := fmt.Sprintf(gistraw, ghun, hash, gidhex)
	logrus.Debugln("[gist]visit url:", u)
	data, err := web.GetData(u)
	if err == nil {
		logrus.Debugln("[gist]get data:", helper.BytesToString(data))
		st, err := strconv.ParseInt(helper.BytesToString(data), 10, 64)
		if err == nil {
			// 600s 内验证成功
			ok := math.Abs(int(time.Now().Unix()-st)) < 600
			if ok {
				_ = db.Insert("member", &member{QQ: qq, Ghun: ghun})
				return true, ""
			}
			return false, "时间戳超时"
		}
		return false, "时间戳格式错误: " + helper.BytesToString(data)
	}
	return false, "无法连接到gist: " + err.Error()
}
