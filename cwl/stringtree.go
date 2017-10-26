
package cwl


type StringTree struct {
  items []string
  itemSeparator string
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
