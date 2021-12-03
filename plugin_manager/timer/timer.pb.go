package timer

type Timer struct {
	Alert                       string `protobuf:"bytes,1,opt"`
	Cron                        string `protobuf:"bytes,2,opt"`
	En1Month4Day5Week3Hour5Min6 int32  `protobuf:"varint,4,opt"`
	Selfid                      int64  `protobuf:"varint,8,opt"`
	Url                         string `protobuf:"bytes,16,opt"`
}

type TimersMap struct {
	Timers map[string]*Timer `protobuf:"bytes,1,rep" protobuf_key:"bytes,1,opt" protobuf_val:"bytes,2,opt"`
}
