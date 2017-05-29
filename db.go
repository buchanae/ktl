package whisk

type database interface {
  get(whiskID string) (string, bool)
  put(whiskID, taskID string)
  list() []string
}

// Mock in-memory database mapping whisk IDs to task IDs
type memdb map[string]string
func (m memdb) get(whiskID string) (string, bool) {
  s, b := m[whiskID]
  return s, b
}
func (m memdb) put(whiskID, taskID string) {
  m[whiskID] = taskID
}
func (m memdb) list() []string {
  var ids []string
  for id, _ := range m {
    ids = append(ids, id)
  }
  return ids
}
