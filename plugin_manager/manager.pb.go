package manager

type Config struct {
	Checkin map[int64]bool   `protobuf:"bytes,1,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,2,opt"`
	Welcome map[int64]string `protobuf:"bytes,2,rep" protobuf_key:"varint,1,opt" protobuf_val:"bytes,2,opt"`
}
