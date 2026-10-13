package control

import (
	"sort"

	ctrl "github.com/FloatTech/zbpctrl"
	zero "github.com/wdvxdr1123/ZeroBot"
)

// ForEachByPrio iterates through managers by their priority.
func ForEachByPrio(iterator func(i int, manager *ctrl.Control[*zero.Ctx]) bool) {
	for i, v := range cpmp2lstbyprio() {
		if !iterator(i, v) {
			return
		}
	}
}

func cpmp2lstbyprio() []*ctrl.Control[*zero.Ctx] {
	managers.RLock()
	defer managers.RUnlock()
	ret := make([]*ctrl.Control[*zero.Ctx], 0, len(managers.M))
	for _, v := range managers.M {
		ret = append(ret, v)
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return enmap[ret[i].Service].prio < enmap[ret[j].Service].prio
	})
	return ret
}
