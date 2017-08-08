/*
  ktl sounds like "kettle"
*/
package main

import (
  "context"
  "fmt"
  "strings"
  "path/filepath"
  "sort"
  "sync"
  "crypto/md5"
  "os"
  "io"
  "github.com/ivaxer/go-xattr"
  "github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
)

func main() {
  s := &localStorage{base: "teststorage"}

  /*
  cat teststorage/input.txt | sort | uniq | md5 > teststorage/md5ed
  */

  nodes := []*node{
    fileNode(s, "input.txt"),
    alpineNode(s, "sort-1", "sort", "input.txt"),
    alpineNode(s, "uniq-1", "uniq", "sort-1"),
    alpineNode(s, "md5-1", "md5sum", "uniq-1"),
    alpineNode(s, "md5-2", "md5sum", "uniq-1"),
    alpineNode(s, "md5-3", "md5sum", "uniq-1"),
    alpineNode(s, "md5-4", "md5sum", "uniq-1"),
    alpineNode(s, "md5-5", "md5sum", "uniq-1"),
    alpineNode(s, "md5-6", "md5sum", "uniq-1"),
    alpineNode(s, "md5-7", "md5sum", "uniq-1"),
    alpineNode(s, "final", "md5sum", "md5-1", "md5-2", "md5-3", "md5-4", "md5-5", "md5-6", "md5-7"),
  }

  cli, err := newTaskClient()
  if err != nil {
    panic(err)
  }

  r := resolver{
    //dryrun: true,
    //cache: &testcache{ data: make(map[string]string) },
    cache: &localCache{s: s},
    executor: testexecutor{cli},
    graph: buildGraph(nodes),
  }

  for i := 0; i < 3; i++ {
    err := r.Resolve("final")
    fmt.Println("Resolve err: ", err)
    fmt.Println("=======================")
  }
}

type testexecutor struct {
  client TaskServiceClient
}
func (t testexecutor) exec(n *node) error {
  task := n.task()
	mar := jsonpb.Marshaler{
		EmitDefaults: true,
		Indent:       "  ",
	}
  s, _ := mar.MarshalToString(task)
  r, err := t.client.CreateTask(context.Background(), task)
  if err != nil {
    panic(err)
  }
  fmt.Println("Exec: ", n.name, s, r)

  // TODO next. wait for task to finish.
  //      want to have these things easily available from funnel.
  return nil
}

func alpineNode(s *localStorage, name, cmd string, in ...string) *node {
  return &node{
    name: name,
    inputs: in,
    task: func() *Task {
      t := Task{
        Name: name,
        Executors: []*Executor{
          {
            ImageName: "alpine",
            Cmd: []string{cmd},
            Stdout: "/w/stdout",
            Workdir: "/w",
          },
        },
        Outputs: []*TaskParameter{
          {
            Url: s.path(name),
            Path: "/w/stdout",
          },
        },
      }

      for _, i := range in {
        p := "/data/" + i
        t.Executors[0].Cmd  = append(t.Executors[0].Cmd, p)
        t.Inputs = append(t.Inputs, &TaskParameter{
          Url: s.path(i),
          Path: p,
        })
      }

      return &t
    },
    taskhash: func() string {
      return cmd
    },
    hash: func() string {
      return s.hash(name)
    },
  }
}

func fileNode(s storage, key string) *node {
  return &node{
    name: key,
    task: func() *Task { return nil },
    taskhash: func() string {
      return "file://" + key
    },
    hash: func() string {
      return s.hash(key)
    },
  }
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





type storage interface {
  hash(key string) string
  path(key string) string
}

type localStorage struct {
  base string
  mtx sync.Mutex
}

func (s *localStorage) path(key string) string {
  p, _ := filepath.Abs(filepath.Join(s.base, key))
  return p
}

func (s *localStorage) hash(key string) string {
  s.mtx.Lock()
  defer s.mtx.Unlock()

  f, err := os.Open(filepath.Join(s.base, key))
	if err != nil {
    return ""
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
    return ""
	}

  return string(h.Sum(nil))
}

type localCache struct {
  s *localStorage
  mtx sync.Mutex
}
func (t *localCache) store(k, h string) {
  t.mtx.Lock()
  defer t.mtx.Unlock()

  p := t.s.path(k)
  xattr.Set(p, "ktl-hash", []byte(h))
}
func (t *localCache) isCached(k, h string) bool {
  t.mtx.Lock()
  defer t.mtx.Unlock()

  p := t.s.path(k)
  if b, err := xattr.Get(p, "ktl-hash"); err == nil {
    return string(b) == h
  }
  return false
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
  task func() *Task
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

  // During dry-run mode, the hash might be empty.
  if hash != "" {
    fmt.Println("store: ", k, hash)
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

func newTaskClient() (TaskServiceClient, error) {
	conn, err := grpc.Dial(
    "localhost:9090",
		grpc.WithInsecure(),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
  return NewTaskServiceClient(conn), nil
}
