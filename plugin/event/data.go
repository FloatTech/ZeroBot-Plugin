package event

type storage int64

// 申请
func (s *storage) setapply(on bool) {
	if on {
		*s |= 0b001
	} else {
		*s &= 0b110
	}
}

// 邀请
func (s *storage) setinvite(on bool) {
	if on {
		*s |= 0b010
	} else {
		*s &= 0b101
	}
}

// 主人
func (s *storage) setmaster(on bool) {
	if on {
		*s |= 0b100
	} else {
		*s &= 0b011
	}
}

// 申请
func (s *storage) isapplyon() bool {
	return *s&0b001 > 0
}

// 邀请
func (s *storage) isinviteon() bool {
	return *s&0b010 > 0
}

// 主人
func (s *storage) ismasteroff() bool {
	return *s&0b100 > 0
}
