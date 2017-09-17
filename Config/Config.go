package config

import (
  "encoding/json"
  "strings"
  "io/ioutil"
  "os"
  "os/exec"
  "bufio"
  "flag"
  "fmt"
)

type Config struct {
	Path        string `json:"path"`
	File        string `json:"file"`
	Integration struct {
		Enable bool `json:"enable"`
		C4     bool `json:"c4"`
	} `json:"integration"`
	Help bool `json:"help"`
}

func Init() (Config) {
  fmt.Println("config.init")

  var defaultPath string = "C:\\Steam\\steamapps\\common\\Counter-Strike Global Offensive\\csgo"
  var defaultFile string = "report.cfg"

  path := flag.String("path", defaultPath, "CSGO Folder")
  file := flag.String("file", defaultFile, "cfg File")
  flag.Parse()

  if len(os.Args) < 2 { // keine Parameter
    if _, err := os.Stat("./config.json"); os.IsNotExist(err) {
      Clear()

      for {
        *path = readCLI("CSGO Path\ndefault: '"+defaultPath+"'\n")
        if *path != "" {

          // val[len(val)-1] lösche den letzten \ falls vorhanden
          // p := (*path)
          // fmt.Println(p[len(p)-1])

          ExistsFolder(*path)
        } else {
          ExistsFolder(defaultPath)
          *path = defaultPath
        }

        b,_ := ExistsFolder(*path)
        if b {
          break
        } else {
          Clear()
          fmt.Println("Ordner nicht gefunden\n")
        }

      }
      Clear()
      *file = readCLI("Report File\ndefault: 'report.cfg'\n")
      if *file == "" {
        *file = "report.cfg"
      }

      return SaveConfigFile(*path, *file)

    } else {
      return LoadConfigFile()
    }

  } else { // parameter gefunden
    return ParseConfig(*path,*file)
  }

  panic("Error")
}


func Clear()  {
  // cmd := exec.Cmd // TODO https://stackoverflow.com/questions/24512112/how-to-print-struct-variables-in-console
  // if runtime.GOOS == "windows" {
    cmd := exec.Command("cmd", "/c", "cls")
  // } else {
  //   cmd = exec.Command("cmd", "/c", "clear")
  // }
  cmd.Stdout = os.Stdout
  cmd.Run()
}



func ExistsFolder(path string) (bool, error) {
  _, err := os.Stat(path)
  if err == nil {return true, nil}
  if os.IsNotExist(err) {return false, nil}
  return true, err
}



func readCLI(txt string) string {
  reader := bufio.NewReader(os.Stdin)
  fmt.Print(txt)

  defaultPathbyte, _, e := reader.ReadLine()
  defaultPath := fmt.Sprintf("%s", defaultPathbyte)
  if e != nil {
    fmt.Printf("error: %#v\n", e)
    os.Exit(1)
  }
  return defaultPath
}



func LoadConfigFile() (Config) {

  file, e := ioutil.ReadFile("./config.json")
  if e != nil {
    fmt.Printf("File error: %v\n", e)
    os.Exit(1)
  }

  var jsontype Config
  err := json.Unmarshal(file, &jsontype)
  if err != nil {
   fmt.Println("error:", err)
  }
  return jsontype
}



func SaveConfigFile(dir, file string) (Config) {
  var jsonBlob = json.RawMessage(`
    {
      "path": "`+strings.Replace(dir, "\\", "\\\\", -1)+`",
      "file": "`+file+`",
      "Integration": {
        "enable": true,
        "c4": true
      },
      "help": true
    }
  `)

  bytes, err := json.Marshal(jsonBlob)
  if err != nil {
    panic("Config konnte nicht gespeichert werden (json error)")
  } else {

    err = ioutil.WriteFile("./config.json", bytes, 0644)
    if err == nil {
      fmt.Println("gespeichert.")

      var jsontype Config
      err := json.Unmarshal(jsonBlob, &jsontype)
      if err != nil {
        panic(err)
      }

      return jsontype
    } else {
      panic(err)
    }

  }
}



func ParseConfig(dir, file string) (Config) {
  var jsonBlob = json.RawMessage(`
    {
      "path": "`+strings.Replace(dir, "\\", "\\\\", -1)+`",
      "file": "`+file+`",
      "Integration": {
        "enable": 1,
        "c4": 1
      }
    }
  `)

  var jsontype Config
  err := json.Unmarshal(jsonBlob, &jsontype)
  if err != nil {
    panic(err)
  }
  return jsontype
}
