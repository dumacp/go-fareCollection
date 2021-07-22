package fare

import "strings"

type prefix []string

func (p prefix) Len() int {
	return len(p)
}

func (p prefix) Less(i, j int) bool {
	leni := strings.Count(p[i], "-")
	lenj := strings.Count(p[j], "-")

	return leni < lenj
}

func (p prefix) Swap(i, j int) {
	temp := p[j]
	p[j] = p[i]
	p[i] = temp
}
