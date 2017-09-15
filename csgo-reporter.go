package main
import (
  "fmt"
  "encoding/json"
  "io/ioutil"
  // "runtime"
  "os"
  "os/exec"
  "bufio"
  "strings"
  "flag"
  "time"
  "log"
  "strconv"
  "net"
  "github.com/tatsushid/go-fastping"
)

type Config struct {
  Path  string `json:"path"`
  File  string `json:"file"`
}

var timer *time.Timer;
var timerempty *time.Timer;

var path, file string
var file_is_empty bool = true
var dmgstatus bool = false
var multiplayer bool = false

var lastplayer string = ""
var dmgs string

func main() {
  path, file := startup()
  clear()
  fmt.Println("add Steam Parameter: -condebug +bind \",\" report.cfg")

  	// Leeren der console.log
  	err := ioutil.WriteFile(path+"\\console.log", []byte("\n"), 0644)
  	if err != nil { panic(err) }

  	writefile(path,file,"")
  	//text := "Damage Taken from \"Loading\x83?\xdd \x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^] 99? Ɵ\"\" - 52 in 2 hits"
  	//fmt.Println(strconv.Quote(line)) // Ascii
  	// https://msdn.microsoft.com/en-us/powershell/reference/5.1/microsoft.powershell.management/get-content
  	// TODO: bugfix für die Symbole
  		//cmd := exec.Command("powershell.exe", "Get-Content", "'"+path+"\\console.log'","-Wait","-Encoding Ascii")
  		//fmt.Println(strconv.Quote(line)) // Ascii
  	// cmd := exec.Command("powershell.exe", "Get-Content", "'"+path+"\\console.log'","-Wait","-Encoding UTF8")
  	cmd := exec.Command("powershell.exe", "Get-Content", "'"+path+"\\console.log'","-Wait","-Encoding Oem")
    	stdout, err := cmd.StdoutPipe()
    	if err != nil { log.Fatal(err) }
    	if err := cmd.Start(); err != nil { log.Fatal(err) }

      scanner := bufio.NewScanner(stdout)
      for scanner.Scan() {
        line := scanner.Text()

  		if strings.Contains(line, "Connected to ") { // Pingt den Server beim verbinden an
  			f := strings.Index(line, "Connected to ");
  			l := strings.LastIndex(line, ":");
  			go ping(line[f+13:l])
  			// ts.spddl.de:2883
  		}

  		if line == "-------------------------" {
  			if dmgstatus {
  				dmgstatus = false
  				// TODO write file
  				//log.Print("lastplayer: ",lastplayer," ",multiplayer)
  				//log.Print("dmgs: ",dmgs," ",multiplayer)

  				if multiplayer {
  					writefile(path,file,dmgs)
  				} else {
  					if lastplayer != "" {
  						writefile(path,file,lastplayer)
  					}
  				}
  				multiplayer = false
  				lastplayer = ""
  				dmgs = ""

  			} else {
  				dmgstatus = true
  			}
  		}

  		if line == "SignalXWriteOpportunity(3)" { // New Round
        fmt.Println(" ")
        log.Print(" ")
  			fmt.Println(" ")
  			//go writedmg("")
  		}

  		if dmgstatus {
  			if strings.Contains(line, "Damage Given to ") && strings.Contains(line, " in ") && strings.Contains(line, " hit") {
  				fmt.Println("\x1b[32;1m",line,"\x1b[0m")
  				go writedmg(line)
  			}
  		} else {
  			if strings.Contains(line, "Damage Taken from ") && strings.Contains(line, " in ") && strings.Contains(line, " hit") {
  				fmt.Println("\x1b[31;1m",line,"\x1b[0m")
  			}
  		}
  		// log.Printf("DEBUG: ",line)
      }

      if err := scanner.Err(); err != nil { log.Fatal(err) }
      if err := cmd.Wait(); err != nil { log.Fatal(err) }

}

func startup() (string, string) {
  var defaultPath string = "C:\\Steam\\steamapps\\common\\Counter-Strike Global Offensive\\csgo"
  var defaultFile string = "report.cfg"

  path := flag.String("path", defaultPath, "CSGO Folder")
  file := flag.String("cfg", defaultFile, "cfg File")
  flag.Parse()

  if len(os.Args) < 2 { // keine Parameter
    if _, err := os.Stat("./config.json"); os.IsNotExist(err) {
      clear()

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
          // TODO Clear
          fmt.Println("Ordner gefunden",*path)
          break
        } else {
          clear()
          fmt.Println("Ordner nicht gefunden\n")
        }

      }
      clear()
      *file = readCLI("Report File\ndefault: 'report.cfg'\n")
      if *file == "" {
        *file = "report.cfg"
      }

      SaveConfigFile(*path, *file)
      return *path,*file

    } else {
      path, file := LoadConfigFile()
      return path,file
    }

  } else { // parameter gefunden
    fmt.Println("Parameter gefunden",path,file)
    return *path,*file
  }

  return "err","err"
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

func LoadConfigFile() (string, string) {
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

  return jsontype.Path, jsontype.File
}

func SaveConfigFile(dir, file string) {
  var jsonBlob = json.RawMessage(`
    {
      "path": "`+strings.Replace(dir, "\\", "\\\\", -1)+`",
      "file": "`+file+`"
    }
  `)

  fmt.Println(jsonBlob)
  bytes, err := json.Marshal(jsonBlob)
  if err != nil {
    fmt.Println("Config konnte nicht gespeichert werden (json error)")
  } else {
    err = ioutil.WriteFile("./config.json", bytes, 0644)
    if err == nil {
      clear()
      fmt.Println("gespeichert.")
    }
  }

}

func clear()  {
  // cmd := exec.Cmd // TODO https://stackoverflow.com/questions/24512112/how-to-print-struct-variables-in-console
  // if runtime.GOOS == "windows" {
    cmd := exec.Command("cmd", "/c", "cls")
  // } else {
  //   cmd = exec.Command("cmd", "/c", "clear")
  // }
  cmd.Stdout = os.Stdout
  cmd.Run()
}


//////////////////////


func writefile(path,file,val string) {
	if val == "" {
		err := ioutil.WriteFile(path+"\\cfg\\"+file, []byte("echo byspddl"), 0644)
		if err != nil { panic(err) }
		file_is_empty = true
	} else {

		if timerempty != nil {
			timerempty.Stop();
		}

		// log.Print("say_team: ",val)
		err := ioutil.WriteFile(path+"\\cfg\\"+file, []byte("say_team "+val), 0644)
		if err != nil { panic(err) }
		file_is_empty = false

		lastplayer = ""
		dmgs = ""

		if timerempty != nil { timerempty.Stop() }
			timerempty = time.AfterFunc(30*time.Second, func() {
			if file_is_empty != true {
				err := ioutil.WriteFile(path+"\\cfg\\"+file, []byte(" "), 0644)
				if err != nil { panic(err) }
			}
			file_is_empty = true
		})
	}



}

func writedmg(val string) {
	// Parse die HP
    f := strings.Index(val, " - "); l := strings.LastIndex(val, " in "); hp, _ := strconv.Atoi(val[f+3:l])

	// TODO wenn 2 gegner über 100 hp verloren haben und es keinen anderen gibt sollte die Datei leeer sein
    if hp <= 99 { // Nur wenn der Gegner noch nicht Tot ist
		//log.Print("writedmg: ",val," (",hp,")")

		if lastplayer != "" {
			dmgs = dmgs+" - "+parsedmg(val)
			multiplayer = true
		} else {
			multiplayer = false
			dmgs = parsedmg(val)
		}
		lastplayer = val
    }
}


func parsedmg(line string) (string){
	_firstPlayer := strings.Index(line, "\"");
	_lastPlayer := strings.LastIndex(line, "\"");
	player := line[_firstPlayer+1:_lastPlayer]

	_firstDmg := strings.Index(line, " - ");
	dmg := line[_firstDmg+3:]
	return player+" "+dmg
}





func ping(ip string) {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		fmt.Println(err)
		return;
	}
		p.AddIPAddr(ra)
		p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("\x1b[36;1mIP: %s, RTT: %v\x1b[0m\n", addr.String(), rtt) //  Round trip time https://de.wikipedia.org/wiki/Ping_(Daten%C3%BCbertragung)
	}
	// p.OnIdle = func() { fmt.Println(ip,"finish") }
	err = p.Run()
	if err != nil {
		// fmt.Println(err)
	}
}











type Ticker interface {
	Duration() time.Duration
	Tick()
	Stop()
}

type ticker struct {
	*time.Ticker
	d time.Duration
}

func (t *ticker) Tick()                   { <-t.C }
func (t *ticker) Duration() time.Duration { return t.d }

func NewTicker(d time.Duration) Ticker {
	return &ticker{time.NewTicker(d), d}
}

type TickFunc func(d time.Duration)

func Countdown(ticker Ticker, duration time.Duration) chan time.Duration {
	remainingCh := make(chan time.Duration, 1)
	go func(ticker Ticker, dur time.Duration, remainingCh chan time.Duration) {
		for remaining := duration; remaining >= 0; remaining -= ticker.Duration() {
			remainingCh <- remaining
			ticker.Tick()
		}
		ticker.Stop()
		close(remainingCh)
	}(ticker, duration, remainingCh)
	return remainingCh
}

func c4timer() {
  // TODO: in einer Var speichern um den Timer zu stopen
  for d := range Countdown(NewTicker(5*time.Second), 40*time.Second) {
    fmt.Println(d)
  }
}
