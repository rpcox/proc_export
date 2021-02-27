// HTTP delivery of /proc walk metrics
package main

import (
  "bytes"
  "bufio"
  //"encoding/json"
  "flag"
//  "fmt"
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


// Set the metric TYPE statement.  All of the metrics will have
// the prefix 'proc_stat'
func setMetricType(r []byte, m string, t string) []byte {
  r = append(r, '#',' ','T','Y','P','E',' ','p','r','o','c','_','s','t','a','t','_')
  line := m + " " + t + "\n"
  return append(r, line...)
}

func setMetric(response []byte, name string, metric []byte) []byte {
  metric_prefix := []byte{'p','r','o','c','_','s','t','a','t','_'}
  response = append(append(append(response, metric_prefix[:]...), name...), ' ')
  response = append(response, metric...)
  return append(response, 10)
}

// Set a particular metric to identify if the monitored process is active
// process running = 1, process not running = 0
func procStatus(response []byte, proc string, status string) []byte {
  response = append(response, '#',' ','T','Y','P','E',' ','p','r','o','c','_','s','t','a','t','_')
  line := proc + "_up gauge\n"
  response = append(response, line...)
  response = append(response, 'p','r','o','c','_','s','t','a','t','_')
  line = proc + "_up " + status + "\n"
  return append(response, line...)
}


//
func getStats(proc_map map[string]string) []byte {
  var response []byte

  for k, path := range proc_map {
    content, err := ioutil.ReadFile(path)

    if err == nil {
      response = procStatus(response, k, "1")
      fields := bytes.Split(content, []byte(" "))
      response = setMetricType(response, "minflt", "gauge")
      response = setMetric(response, "minflt", fields[9])
      response = setMetricType(response, "cminflt", "gauge")
      response = setMetric(response, "cminflt", fields[10])
      response = setMetricType(response, "majflt", "gauge")
      response = setMetric(response, "majflt", fields[11])
      response = setMetricType(response, "cmajflt", "gauge")
      response = setMetric(response, "cmajflt", fields[12])
      response = setMetricType(response, "utime_seconds", "counter")
      response = setMetric(response, "utime_seconds", fields[13])
      response = setMetricType(response, "stime_seconds", "counter")
      response = setMetric(response, "stime_seconds", fields[14])
      response = setMetricType(response, "nice", "gauge")
      response = setMetric(response, "nice", fields[18])
      response = setMetricType(response, "num_threads", "gauge")
      response = setMetric(response, "num_threads", fields[19])
      response = setMetricType(response, "vsize_bytes", "gauge")
      response = setMetric(response, "vsize_bytes", fields[22])
      response = setMetricType(response, "rss_pages", "gauge")
      response = setMetric(response, "rss_pages", fields[23])
      response = setMetricType(response, "rt_priority", "gauge")
      response = setMetric(response, "rt_priority", fields[39])

    } else {
      response = procStatus(response, k, "0")
      delete (proc_map, k)
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

