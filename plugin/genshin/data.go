package genshin

type storage uint64

func (s *storage) is5starsmode() bool {
	return *s&1 == 1
}

func (s *storage) setmode(is5stars bool) bool {
	if is5stars {
		*s |= 1
	} else {
		*s &= 0xffffffff_fffffffe
	}
	return is5stars
}
