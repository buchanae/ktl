/*
  ktl sounds like "kettle"
*/
package main

import (
  "time"
  "fmt"
  "strings"
  "sort"
  "sync"
)

func main() {
  s := &storage{data: make(map[string]int)}

  /*
    a -
       |
    b ---- d
    |       \
    ----------- f
            /
    c ---- e
  */


  // Leaf nodes (workflow inputs)
  s.set("a", 1)
  //s.set("b", 2)
  s.set("c", 3)

  incamt := new(int)
  *incamt = 1

  nodes := []*node{
    valueNode(s, "a"),
    valueNode(s, "b"),
    valueNode(s, "c"),

    // 1 + 2 = 3
    incTask(s, "b", "a", incamt),
    sumTask(s, "d", "a", "b"),

    // 3 + 1 = 4
    incTask(s, "e", "c", incamt),

    // 2 + 3 + 4 = 9
    sumTask(s, "f", "b", "d", "e"),
  }

  r := resolver{
    dryrun: true,
    cache: &testcache{ data: make(map[string]string) },
    executor: testexecutor{},
    graph: buildGraph(nodes),
  }

  for i := 0; i < 3; i++ {
    err := r.Resolve("f")
    s.set("c", i)
    //s.set("c", 8)
    //*incamt = *incamt + 1
    fmt.Println("Result: ", s.get("f"), err)
    fmt.Println("==================")
  }

  nodes2 := []*node{
    valueNode(s, "f"),
    // f + 1
    incTask(s, "g", "f", incamt),
  }

  r2 := resolver{
    //dryrun: true,
    cache: &testcache{ data: make(map[string]string) },
    executor: testexecutor{},
    graph: buildGraph(nodes2),
  }
  r2.Resolve("g")
  fmt.Println("Result: ", s.get("g"))
  fmt.Println("==================")
}


/*
TODO

- funnel executor
- only run task once
- connect to actual object stores
- persistent cache
- prevalidate graph before enabling execution
- rethink taskhash() and hash() and hashing code in general
- rethink executor and cache organization
*/


// TODO should check that two nodes can't create the same output
//      or build this into the node definition (namespacing?)
type graph map[string]*node
func buildGraph(nodes []*node) graph {
  g := make(graph)
  for _, n := range nodes {
    g[n.name] = n
  }
  return g
}

func valueNode(s *storage, key string) *node {
  return &node{
    name: key,
    task: func() {},
    taskhash: func() string {
      return "noop()"
    },
    hash: func() string {
      return s.hash(key)
    },
  }
}

func incTask(s *storage, out string, in string, amt *int) *node {
  return &node{
    name: out,
    inputs: []string{in},
    task: func() {
      fmt.Println("INC AMT", *amt)
      s.set(out, s.get(in) + *amt)
      time.Sleep(time.Second * 3)
    },
    taskhash: func() string {
      return fmt.Sprintf("inc(%d)", *amt)
    },
    hash: func() string {
      return s.hash(out)
    },
  }
}

func sumTask(s *storage, out string, keys ...string) *node {
  return &node{
    name: out,
    inputs: keys,
    task: func() {
      t := 0
      for _, k := range keys {
        t += s.get(k)
      }
      s.set(out, t)
      time.Sleep(time.Second * 3)
    },
    taskhash: func() string {
      return "sum()"
    },
    hash: func() string {
      return s.hash(out)
    },
  }
}



type storage struct {
  mtx sync.Mutex
  data map[string]int
}

func (s *storage) hash(key string) string {
  s.mtx.Lock()
  defer s.mtx.Unlock()
  if v, ok := s.data[key]; ok {
    return fmt.Sprintf("%d", v)
  }
  // TODO hack for checking whether the value exists in storage
  fmt.Println("no such key, cannot hash", key)
  return ""
}

func (s *storage) get(key string) int {
  s.mtx.Lock()
  defer s.mtx.Unlock()
  return s.data[key]
}

func (s *storage) set(key string, val int) {
  s.mtx.Lock()
  defer s.mtx.Unlock()
  s.data[key] = val
}



type testexecutor struct {}
func (t testexecutor) exec(n *node) error {
  fmt.Println("Exec: ", n.name)
  n.task()
  return nil
}


type testcache struct {
  mtx sync.Mutex
  data map[string]string
}
func (t *testcache) store(k, h string) {
  t.mtx.Lock()
  defer t.mtx.Unlock()
  t.data[k] = h
}
func (t *testcache) isCached(k, h string) bool {
  t.mtx.Lock()
  defer t.mtx.Unlock()
  if e, ok := t.data[k]; ok {
    return e == h
  }
  return false
}



type node struct {
  name string
  inputs []string
  task func()
  taskhash func() string
  hash func() string
}

type cache interface {
  isCached(k, h string) bool
  store(k, h string)
}

type executor interface {
  exec(n *node) error
}




type resolver struct {
  cache
  executor
  graph graph
  dryrun bool
}

func (r *resolver) Resolve(k string) error {
  n, ok := r.graph[k]
  if !ok {
    return fmt.Errorf("can't find node: %s", k)
  }

  // Resolve inputs in parallel
  inputs, err := r.resolveInputs(n)
  if err != nil {
    return err
  }

  hash := r.hash(k, n.taskhash(), inputs)

  // During dry-run mode, the hash might be empty.
  if hash != "" {
    // Check for a cached value. If the value already exists, there's nothing to do.
    if r.cache.isCached(k, hash) {
      fmt.Println("cached", k)
      return nil
    }
  }

  // TODO consider moving dryrun to executor, and move caching also so that
  //      executor can coordinate both.
  if r.dryrun {
    fmt.Println("dryrun exec", k)
    return nil
  }

  // Execute the node's task to create the value.
  xerr := r.exec(n)
  if xerr != nil {
    return xerr
  }

  fmt.Println("store: ", k, hash)
  // During dry-run mode, the hash might be empty.
  if hash != "" {
    r.cache.store(k, hash)
  }
  return nil
}

func (r *resolver) resolveInputs(n *node) ([]string, error) {
  keys := make([]string, len(n.inputs))
  errs := make([]error, len(n.inputs))
  wg := sync.WaitGroup{}
  wg.Add(len(n.inputs))

  for i, k := range n.inputs {
    go func(i int, k string) {
      err := r.Resolve(k)
      keys[i] = k
      errs[i] = err
      wg.Done()
    }(i, k)
  }
  wg.Wait()

  // TODO need to figure if/how to handle multiple errors
  // TODO should the first error cancel all the other goroutines?
  for _, err := range errs {
    if err != nil {
      return nil, err
    }
  }

  return keys, nil
}

func (r *resolver) hash(k string, task string, inputs []string) string {
  var hashes []string
  for _, k := range inputs {
    h := r.graph[k].hash()
    // If any of the inputs can't be hashed, then the hash can't be built.
    // This is useful for dry-run mode, when the intermediate cache might
    // have missing values. During normal execution, a missing hash is a problem?
    if h == "" {
      return ""
    }
    hashes = append(hashes, h)
  }
  sort.Strings(hashes)
  base := k + "." + task
  if hashes != nil {
    base += "-" + strings.Join(hashes, "-")
  }
  return base
}
