package main

/*
Lessons:
- checking the state of multiple, sequential tasks in parallel would help.
*/

/*
TODO
- command to check the outputs of all nodes, to identify corrupt tasks which are marked as complete
  but have missing outputs
- close connections properly, to reduce error logs
- use one connection and process, to reduce overhead and error logs
- could probably speed this up more by saving the IDs in a single database,
  or some other trick that avoids opening hundreds of files (task content hash?)
  - although, doesn't the OS cache these files?
*/

import (
  "context"
  "strings"
  "path/filepath"
  "io/ioutil"
  "fmt"
  "os"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
  "github.com/spf13/cobra"
  "time"
)

// RootCmd represents the root command
var rootCmd = &cobra.Command{
    Use:           "ktl",
    SilenceErrors: true,
    SilenceUsage:  true,
}

func init() {
  rootCmd.AddCommand(statusCmd)
  rootCmd.AddCommand(runCmd)
}

func main() {
  if err := rootCmd.Execute(); err != nil {
    fmt.Println(err.Error())
    os.Exit(1)
  }
}

func globTasks(dir string) []string {
  var res []string
  m, _ := filepath.Glob(filepath.Join(dir, "task*.json"))
  res = append(res, m...)
  return res
}

var statusCmd = &cobra.Command{
    Use:   "status [task.json ...]",
    RunE: func(cmd *cobra.Command, args []string) error {
        if len(args) == 0 {
            return cmd.Help()
        }

        conn, err := grpc.Dial(
          "funnel_server_1:9090",
          grpc.WithInsecure(),
          //grpc.WithBlock(),
        )
        defer conn.Close()
        if err != nil {
          panic(err)
        }
        cli := NewTaskServiceClient(conn)
        tasks := map[string]task{}
        wg := sync.WaitGroup{}

        // TODO might accidentally pass task files instead of directory
        //      in which case, globTasks will break
        for _, arg := range args {
          for _, path := range globTasks(arg) {
            wg.Add(1)
            go func(path string) {
              defer wg.Done()
              ctx := context.Background()
              id := loadID(path)
              resp, err := cli.GetTask(ctx, &GetTaskRequest{
                Id: id,
                View: TaskView_BASIC,
              })
              if err != nil && !isNotFound(err) {
                panic(err)
              }
              // TODO move this to use an http cache under the hood
              //      so that keeping the map isn't necessary.
              tasks[path] = task{Task: resp, path: path}
            }(path)
          }
        }

        wg.Wait()

        for _ arg := range args {
          var ts []task
          for _, path := range globTasks(arg) {
            if t, ok := tasks[path]; ok {
              ts = append(ts, t)
            }
          }
          doStatus(ts)
        }

        return nil
    },
}


var statusFlags = struct {
  running bool
  queued bool
  unknown bool
  initializing bool
  canceled bool
  complete bool
  err bool
  syserr bool
  anyerr bool
  done bool

  last bool
  latest bool

  base bool
  path bool
  id bool
  state bool
  duration bool
  minDuration time.Duration
}{}


func init() {
  f := statusCmd.Flags()
  f.BoolVar(&statusFlags.running, "running", statusFlags.running, "Running tasks only")
  f.BoolVar(&statusFlags.queued, "queued", statusFlags.queued, "queued tasks only")
  f.BoolVar(&statusFlags.unknown, "unknown", statusFlags.unknown, "unknown tasks only")
  f.BoolVar(&statusFlags.initializing, "initializing", statusFlags.initializing, "initializing tasks only")
  f.BoolVar(&statusFlags.canceled, "canceled", statusFlags.canceled, "canceled tasks only")
  f.BoolVar(&statusFlags.complete, "complete", statusFlags.complete, "complete tasks only")
  f.BoolVar(&statusFlags.err, "err", statusFlags.err, "err tasks only")
  f.BoolVar(&statusFlags.syserr, "syserr", statusFlags.syserr, "syserr tasks only")
  f.BoolVar(&statusFlags.anyerr, "anyerr", statusFlags.anyerr, "anyerr tasks only")
  f.BoolVar(&statusFlags.done, "done", statusFlags.done, "done tasks only")

  f.BoolVar(&statusFlags.last, "last", statusFlags.last, "last tasks only")
  f.BoolVar(&statusFlags.latest, "latest", statusFlags.latest, "Latest tasks only")

  f.BoolVar(&statusFlags.base, "base", statusFlags.base, "include base field")
  f.BoolVar(&statusFlags.path, "path", statusFlags.path, "include path field")
  f.BoolVar(&statusFlags.id, "id", statusFlags.id, "include id field")
  f.BoolVar(&statusFlags.state, "state", statusFlags.state, "include state field")
  f.BoolVar(&statusFlags.duration, "duration", statusFlags.duration, "include duration field")
  f.DurationVar(&statusFlags.minDuration, "min-duration", statusFlags.minDuration, "filter out rows where duration < min-duration")
}

type row struct {
  ID string
  BasePath string
  Path string
  State State
  IsLast bool
  Duration time.Duration
}

func getTask(cli TaskServiceClient) {
}

type task struct {
  *Task
  path string
}

func doStatus(tasks []task) {

  // Default column config
  if !statusFlags.id && !statusFlags.base && !statusFlags.path && !statusFlags.state && !statusFlags.duration {
    statusFlags.id = true
    statusFlags.path = true
    statusFlags.state = true
  }

  if statusFlags.last {
    args = args[len(args) - 1:]
  }

  var rows []*row

  base := filepath.Dir(args[0])

  for i, arg := range args {
    r := row{
      ID: id,
      BasePath: base,
      Path: arg,
      // TODO assumes linear sequence
      IsLast: i == len(args) - 1,
    }
    rows = append(rows, &r)

    if id != "" {
      if err != nil {
        panic(err)
      }

      r.State = resp.GetState()

      if resp.Logs != nil && resp.Logs[0].StartTime != "" {
        start, _ := time.Parse(time.RFC3339, resp.Logs[0].StartTime)

        if resp.Logs[0].EndTime != "" {
          end, _ := time.Parse(time.RFC3339, resp.Logs[0].EndTime)
          r.Duration = end.Sub(start)
        } else {
          r.Duration = time.Since(start)
        }
      }
    }
  }

  var filtered []*row

  for i, r := range rows {
    switch {
    case statusFlags.latest && (r.State == State_UNKNOWN || r.State == State_QUEUED):
    case statusFlags.latest && i < len(rows) - 1 && rows[i + 1].State != State_UNKNOWN:
    default:
      filtered = append(filtered, r)
    }
  }

  for _, r := range filtered {
    var cols []string

    // TODO oops! these are not OR'ed together!
    //      this is the point where you realize a query language is needed
    if statusFlags.running && r.State != State_RUNNING {
      continue
    }
    if statusFlags.complete && r.State != State_COMPLETE {
      continue
    }
    if statusFlags.queued && r.State != State_QUEUED {
      continue
    }
    if statusFlags.unknown && r.State != State_UNKNOWN {
      continue
    }
    if statusFlags.initializing && r.State != State_INITIALIZING {
      continue
    }
    if statusFlags.err && r.State != State_ERROR {
      continue
    }
    if statusFlags.syserr && r.State != State_SYSTEM_ERROR {
      continue
    }
    if statusFlags.anyerr && r.State != State_SYSTEM_ERROR && r.State != State_ERROR {
      continue
    }
    if statusFlags.canceled && r.State != State_CANCELED {
      continue
    }
    if statusFlags.done && !isDone(r.State) {
      continue
    }
    if statusFlags.minDuration > 0 && r.Duration < statusFlags.minDuration {
      continue
    }

    // TODO cols should match the order of flags
    if statusFlags.id {
      cols = append(cols, fmt.Sprintf("%20s", r.ID))
    }
    if statusFlags.base {
      cols = append(cols, r.BasePath)
    }
    if statusFlags.path {
      cols = append(cols, r.Path)
    }
    if statusFlags.state {
      cols = append(cols, r.State.String())
    }
    if statusFlags.duration {
      cols = append(cols, r.Duration.String())
    }

    fmt.Println(strings.Join(cols, "\t"))
  }
}

func isDone(s State) bool {
	return s == State_COMPLETE || s == State_ERROR || s == State_SYSTEM_ERROR ||
		s == State_CANCELED
}

func isNotFound(err error) bool {
  s, ok := status.FromError(err)
  return ok && s.Code() == codes.NotFound
}

func loadID(path string) string {
  b, err := ioutil.ReadFile(path + ".id")
  if os.IsNotExist(err) {
    return ""
  }
  if err != nil {
    panic(err)
  }
  return strings.Trim(string(b), "\n")
}
