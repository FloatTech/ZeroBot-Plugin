package fortune

type Conf struct {
	Kind map[int64]uint32 `protobuf:"bytes,1,rep" protobuf_key:"varint,1,opt" protobuf_val:"varint,0,opt"`
}
