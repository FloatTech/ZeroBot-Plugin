package partygame

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"
	"time"
)

// Session 会话操作
type Session struct {
	GroupID    int64   // 群id
	Creator    int64   // 创建者
	Users      []int64 // 参与者
	Max        int64   // 最大人数
	Cartridges []int   // 弹夹
	IsValid    bool    // 是否有效
	ExpireTime int64   // 过期时间
	CreateTime int64   // 创建时间
}

var dataPath string

var rlmu sync.RWMutex

func checkFile(path string) {
	rlmu.Lock()
	defer rlmu.Unlock()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		_, err := os.Create(path)
		if err != nil {
			return
		}
	}
	dataPath = path
}

func saveItem(dataPath string, info Session) {
	interact := loadSessions(dataPath)
	rlmu.Lock()
	defer rlmu.Unlock()
	if len(interact) == 0 {
		interact = append(interact, info)
	} else {
		for i, v := range interact {
			if v.GroupID == info.GroupID {
				interact[i] = info
				break
			}
		}
	}
	bytes, err := json.Marshal(&interact)
	if err != nil {
		panic(err)
	}
	// 将数据data写入文件filePath中，并且修改文件权限为755
	if err = ioutil.WriteFile(dataPath, bytes, 0644); err != nil {
		panic(err)
	}
}

func loadSessions(dataPath string) []Session {
	// 读取指定文件内容，返回的data是[]byte类型数据
	rlmu.RLock()
	defer rlmu.RUnlock()
	ss := make([]Session, 0)
	data, err := ioutil.ReadFile(dataPath)
	if err != nil {
		return ss
	}
	if err = json.Unmarshal(data, &ss); err != nil {
		return ss
	}
	return ss
}

func getSession(gid int64, dataPath string) Session {
	interact := loadSessions(dataPath)
	for _, v := range interact {
		if v.GroupID == gid {
			return v
		}
	}
	return Session{}
}

// 添加会话
func addSession(gid, uid int64, dataPath string) {
	cls := Session{}
	cls.GroupID = gid
	cls.Creator = uid
	cls.Users = append(cls.Users, uid)
	cls.IsValid = false
	cls.Max = 3
	cls.Cartridges = cls.rotateRoulette()
	cls.ExpireTime = 300
	cls.CreateTime = time.Now().Unix()

	saveItem(dataPath, cls)
}

// 获取参与人数
func (cls Session) countUser() int {
	return len(cls.Users)
}

// 加入会话
func (cls Session) addUser(userID int64) {
	cls.Users = append(cls.Users, userID)
	saveItem(dataPath, cls)
}

// 关闭
func (cls Session) close() {
	interact := loadSessions(dataPath)

	run := make([]Session, 0)
	for _, v := range interact {
		if v.GroupID == cls.GroupID {
			continue
		}
		run = append(run, v)
	}

	bytes, err := json.Marshal(&run)
	if err != nil {
		panic(err)
	}
	// 将数据data写入文件filePath中，并且修改文件权限为755
	if err = ioutil.WriteFile(dataPath, bytes, 0644); err != nil {
		panic(err)
	}
}

// 判断会话是否过期
func (cls Session) isExpire() bool {
	// 当前时间
	now := time.Now().Unix()
	// 创建时间加存活时间
	return cls.CreateTime+cls.ExpireTime < now
}

// 判断是否在队伍中
func (cls Session) checkJoin(uid int64) bool {
	// 判断是否在参与者列表中
	for _, j := range cls.Users {
		if j == uid {
			return true
		}
	}
	return false
}

// 判断是否轮到用户
func (cls Session) checkTurn(uid int64) bool {
	return cls.Users[0] == uid
}

// 剩余子弹数
func (cls Session) cartridgesLeft() int {
	return len(cls.Cartridges)
}

// 开火
func (cls Session) openFire() bool {
	// 压出头部
	bullet := cls.Cartridges[0]
	cls.Cartridges = cls.Cartridges[1:]
	if bullet == 1 {
		return true
	}
	// 获取开枪人
	user := cls.Users[0]
	// 人员轮转
	cls.Users = cls.Users[1:]
	cls.Users = append(cls.Users, user)

	saveItem(dataPath, cls)
	return false
}

// 打乱参与人顺序
func (cls Session) rotateUser() {
	// 随机打乱数组
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cls.Users), func(i, j int) { cls.Users[i], cls.Users[j] = cls.Users[j], cls.Users[i] })
	saveItem(dataPath, cls)
}

// 旋转轮盘
func (cls Session) rotateRoulette() []int {
	// 创建6个仓位的左轮弹夹
	cartridges := []int{1, 0, 0, 0, 0, 0}
	// 随机打乱数组
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cartridges), func(i, j int) { cartridges[i], cartridges[j] = cartridges[j], cartridges[i] })
	return cartridges
}