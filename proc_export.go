// HTTP delivery of /proc walk metrics
package main

import (
  "bufio"
  //"encoding/json"
  "flag"
  //"fmt"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "regexp"
  "strconv"
  "strings"
  //"time"
)


var proc string

func init() {
  flag.StringVar(&proc, "proc",     "", "Comma separated list of process names")
}

// With the list supplied from the -proc flag, use those process names
// as keys in a map and set the values of the map to string nil
//
func makeMap(s string) map[string]string  {
  list := strings.Split(s, ",")
  m := make(map[string]string)
  for _, val := range list {
   m[val] = ""
  }

  return m
}

// With the keys from the process list map find the PID of the process
// then form the path to it's respective /proc/[pid]/stat file to be
// used as the value for the key
//
func findPids(proc_map map[string]string) {
  files, err := ioutil.ReadDir("/proc")
  if err != nil { log.Fatal(err) }

  // See proc(5). The second field in the file is the file name of
  // the executable (a key used in proc_map)
  regex_map := make(map[string]*regexp.Regexp)
  for k, _ := range proc_map {
    regex_map[k] = regexp.MustCompile(k)
  }

  for _, file := range files {
    _, err := strconv.Atoi(file.Name())
    if err != nil { break }
    path := "/proc/" + file.Name() + "/stat"
    f, err := os.Open(path)
    if err != nil { continue }
    r := bufio.NewReader(f)
    line, err := r.ReadString('\n')
    f.Close()

    for k := range regex_map {
      if proc_map[k] == "" && regex_map[k].MatchString(line) {
        proc_map[k] = path
      }
    }
  }
}

//
func getStats(proc_map map[string]string) []byte {
  var response []byte

  for k, path := range proc_map {
    content, err := ioutil.ReadFile(path)

    if err == nil {
      response = append(response, content...)
    } else {
      response = append(response, k + ": failed collection"...)
      delete (proc_map, k)   // the process at this PID is no more
      continue
    }
  }

  return response
}

func main() {
  flag.Parse()
  proc_map := makeMap(proc)
  findPids(proc_map)

  http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
    response := getStats(proc_map)
    w.Header().Set("Content-Type", "application/text")
    w.Write(response)
    })

  log.Fatal(http.ListenAndServe("localhost:9000", nil))
}

