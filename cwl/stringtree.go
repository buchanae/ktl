
package cwl


type StringTree struct {
  items []StringTreeElement
  itemSeparator string
}

type StringTreeElement interface {
  ToArray() []string
}

func NewStringTree() StringTree {
  return StringTree{[]StringTreeElement{}, ""}
}

func String2Tree(x string) StringTree {
  return StringTree{[]string{x}, ""}
}

type StringTreeString string

func (self StringTreeString) ToArray() []string {
  return []string{self}
}

func (self StringTree) ToArray() []string {
  return self.items
}

func (self StringTree) Append(x ...string) StringTree {
  return StringTree{append(self.items, x...), self.itemSeparator}
}

func (self StringTree) Extend(x StringTree) StringTree {
  return StringTree{append(self.items, x.items...), self.itemSeparator}
}


func (self StringTree) SetSeperator(x string) StringTree {
  return StringTree{self.items, x}
}

func array_join(v []string, j string) []string {
  if len(v) < 2 {
    return v
  }
  out := make([]string, (len(v) * 2) - 1)
  for i := range v {
    out = append(out, v[i])
    if i < len(v) - 1 {
      out = append(out, j)
    }
  }
  return out
}
