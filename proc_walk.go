// Walk /proc looking for a process by name
// go run proc_walk.go -proc process_name_1,process_name_2
// go run proc_walk.go -proc pluma,gnome-calc
package main


import (
  "bufio"
  "flag"
  "fmt"
  "io/ioutil"
  "log"
  "os"
  "regexp"
  "strconv"
  "strings"
)


var proc string

func init() {
  flag.StringVar(&proc, "proc",     "", "Comma separated list of process names")
}

func make_map(s string) map[string]int  {
  list := strings.Split(s, ",")
  m := make(map[string]int)
  for _, val := range list {
   m[val] = 0
  }

  return m
}


func find_pids(proc_list map[string]int) {
  files, err := ioutil.ReadDir("/proc")
  if err != nil { log.Fatal(err) }
  regex_map := make(map[string]*regexp.Regexp)
  for k, _ := range proc_list {
    regex_map[k] = regexp.MustCompile(k)
  }

  for _, file := range files {
    pid, err := strconv.Atoi(file.Name())
    if err != nil { break }               // we're into text and done with PIDs

    f, err := os.Open("/proc/" + file.Name() + "/stat")
    if err != nil { continue }            // 1M reason's it might not be there. move on
    r := bufio.NewReader(f)
    line, err := r.ReadString('\n')
    f.Close()                             // don't defer, close now
    for k := range regex_map {
      if proc_list[k] == 0 && regex_map[k].MatchString(line) {
        proc_list[k] = pid
      }
    }
  }
}


func main() {
  flag.Parse()
  proc_list := make_map(proc)
  find_pids(proc_list)

  for k, v := range proc_list {
    fmt.Println(k,v)
  }
}

