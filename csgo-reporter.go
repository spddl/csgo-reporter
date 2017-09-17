package main

import (
  "fmt"
  // "runtime"
  "os/exec"
  "io/ioutil"
  "bufio"
  "strings"
  "time"
  "log"
  "strconv"
  "net"
  "github.com/tatsushid/go-fastping"
  "./Config"
  "./GameStateIntegration"
)


var path, file string
var file_is_empty bool = true
var dmgstatus bool = false
var multiplayer bool = false

var lastplayer string = ""
var dmgs string


func main() {
  Config := config.Init()

  config.Clear()
  if Config.Help {
    fmt.Println("\x1b[36;1madd Steam Parameter: -condebug +bind \",\" report.cfg\x1b[0m")
  }

  if Config.Integration.Enable {
    go GameStateIntegration.Start(Config.Path) // https://github.com/dank/go-csgsi
  }

   	// Leeren der console.log
  	err := ioutil.WriteFile(Config.Path+"\\console.log", []byte("\n"), 0644)
  	if err != nil { panic(err) }

  	writefile(Config.Path,Config.File,"")
  	//text := "Damage Taken from \"Loading\x83?\xdd \x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^] 99? Ɵ\"\" - 52 in 2 hits"
  	//fmt.Println(strconv.Quote(line)) // Ascii
  	// https://msdn.microsoft.com/en-us/powershell/reference/5.1/microsoft.powershell.management/get-content
  	// TODO: bugfix für die Symbole
  		//cmd := exec.Command("powershell.exe", "Get-Content", "'"+path+"\\console.log'","-Wait","-Encoding Ascii")
  		//fmt.Println(strconv.Quote(line)) // Ascii
  	// cmd := exec.Command("powershell.exe", "Get-Content", "'"+path+"\\console.log'","-Wait","-Encoding UTF8")
      cmd := exec.Command("powershell.exe", "Get-Content", "'"+Config.Path+"\\console.log'","-Wait","-Encoding Oem")
    	stdout, err := cmd.StdoutPipe()
    	if err != nil { log.Fatal(err) }
    	if err := cmd.Start(); err != nil { log.Fatal(err) }

      scanner := bufio.NewScanner(stdout)
      for scanner.Scan() {
      line := scanner.Text()

		if strings.Contains(line, "Confirmed best official datacenter ping: ") { // errechnete Verzögung von Valve
			log.Print("\x1b[36;1m",line,"\x1b[0m")
  	}

  		if strings.Contains(line, "Connected to ") { // Pingt den Server beim verbinden an
  			f := strings.Index(line, "Connected to ");
  			l := strings.LastIndex(line, ":");
  			if f == 0 {
  				go ping(line[f+13:l])
  			}
  		}

  		if line == "-------------------------" {
  			if dmgstatus {
  				dmgstatus = false
  				// TODO write file
  				//log.Print("lastplayer: ",lastplayer," ",multiplayer)
  				//log.Print("dmgs: ",dmgs," ",multiplayer)

  				if multiplayer {
  					writefile(Config.Path,Config.File,dmgs)
  				} else {
  					if lastplayer != "" {
  						writefile(Config.Path,Config.File,lastplayer)
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
  			// fmt.Println(" ")
  			// log.Print(" ")
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

var timerempty *time.Timer;

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
				err := ioutil.WriteFile(path+"\\cfg\\"+file, []byte("echo byspddl"), 0644)
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
