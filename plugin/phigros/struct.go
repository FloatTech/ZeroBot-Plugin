package phigros

// 资源路径
const (
	// 课题模式图标
	Challengemode = "challengemode/"
	// 字体
	Font = "font/font.ttf"
	// 评级
	Rank = "rank/"
	// 曲绘
	Illustration = "illustration/"
	// 图标
	Icon = "icon.png"
)

/*const (
	x,y float64 = 188,682
)*/

/*var (
	// 排名背景
	x, y float64 = 188, 682
	w, h float64 = 70, 44
	// 图片
	x1, y1 float64 = 256, 682
	w1, h1 float64 = 346, 238
	// 定数背景
	x2, y2 float64 = 152, 821
	w2, h2 float64 = 138, 94
	// 分数背景
	x3, y3 float64 = 596, 694
	w3, h3 float64 = 518, 218
	// 边缘
	x4, y4 float64 = 1114, 692
	w4, h4 float64 = 6, 222
	// 真图片
	x5, y5 int = 194, 682
	// 排名
	x6, y6 float64 = 178, 714
	// 分数线
	x7, y7 float64 = 724, 824
	w7, h7 float64 = 326, 2
)
var (
	//level
	x8, y8 float64 = 144, 856
	//level2
	x9, y9 float64 = 138, 898
	// rank
	x10, y10 int = 600, 770
	// score
	x11, y11 float64 = 720, 798
	// name
	x12, y12 float64 = 596, 740
	// acc
	x13, y13 float64 = 724, 878
)*/

// 角度
// const a float64 = 75

var checkchall = map[string]int64{
	"rainbow": 5,
	"gold":    4,
	"red":     3,
	"blue":    2,
	"green":   1,
}

var checkdiff = map[string][]int{
	"AT": {56, 56, 56, 255},
	"IN": {190, 45, 35, 255},
	"HD": {3, 115, 190, 255},
	"EZ": {15, 180, 145, 255},
}
