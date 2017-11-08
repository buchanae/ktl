package cwl

type StringTree struct {
	items         []StringTreeElement
	itemSeparator string
}

type StringTreeElement interface {
	ToArray() []string
}

func NewStringTree() StringTree {
	return StringTree{[]StringTreeElement{}, ""}
}

func String2Tree(x string) StringTree {
	return StringTree{[]StringTreeElement{StringTreeString(x)}, ""}
}

type StringTreeString string

func (self StringTreeString) ToArray() []string {
	return []string{string(self)}
}

func (self StringTree) ToArray() []string {
	o := [][]string{}
	for _, i := range self.items {
		o = append(o, i.ToArray())
	}
	return array_array_join(o, self.itemSeparator)
}

func (self StringTree) Append(x ...string) StringTree {
	o := []StringTreeElement{}
	for _, i := range x {
		o = append(o, StringTreeString(i))
	}
	return StringTree{append(self.items, o...), self.itemSeparator}
}

func (self StringTree) Extend(x StringTree) StringTree {
	return StringTree{append(self.items, x.items...), self.itemSeparator}
}

func (self StringTree) SetSeperator(x string) StringTree {
	return StringTree{self.items, x}
}

func array_array_join(v [][]string, j string) []string {
	if len(v) == 0 {
		return []string{}
	}
	out := []string{}
	for x := range v {
		for y := range v[x] {
			out = append(out, v[x][y])
		}
		if len(j) > 0 && x < len(v)-1 {
			out = append(out, j)
		}
	}
	return out
}

func array_join(v []string, j string) []string {
	if len(v) < 2 {
		return v
	}
	out := make([]string, (len(v)*2)-1)
	for i := range v {
		out = append(out, v[i])
		if len(j) > 0 && i < len(v)-1 {
			out = append(out, j)
		}
	}
	return out
}
