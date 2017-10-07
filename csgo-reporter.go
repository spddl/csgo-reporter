package main

import (
	"fmt"
	// "runtime"
	"bufio"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"./Config"
	"./GameStateIntegration"
	"./SmurfChecker"

	"github.com/tatsushid/go-fastping"
)

var path, file string
var fileIsEmpty = true
var dmgstatus = false
var multiplayer = false
var SmurfCheckerStatus = false

var lastplayer string
var dmgs string

var Array []string

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
	if err != nil {
		panic(err)
	}

	writefile(Config.Path, Config.File, "")
	//text := "Damage Taken from \"Loading\x83?\xdd \x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^\x83-^] 99? Ɵ\"\" - 52 in 2 hits"
	//fmt.Println(strconv.Quote(line)) // Ascii
	// https://msdn.microsoft.com/en-us/powershell/reference/5.1/microsoft.powershell.management/get-content
	// TODO: bugfix für die Symbole
	//cmd := exec.Command("powershell.exe", "Get-Content", "'"+path+"\\console.log'","-Wait","-Encoding Ascii")
	//fmt.Println(strconv.Quote(line)) // Ascii
	// cmd := exec.Command("powershell.exe", "Get-Content", "'"+path+"\\console.log'","-Wait","-Encoding UTF8")
	cmd := exec.Command("powershell.exe", "Get-Content", "'"+Config.Path+"\\console.log'", "-Wait", "-Encoding Oem")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Confirmed best official datacenter ping: ") { // errechnete Verzögung von Valve
			log.Print("\x1b[36;1m", line, "\x1b[0m")
		}

		if strings.Index(line, "udp/ip  : ") == 0 {
			l := strings.LastIndex(line, ":")
			go ping(line[10:l])
		}
		if strings.Contains(line, "Connected to ") { // Pingt den Server beim verbinden an
			f := strings.Index(line, "Connected to ")
			l := strings.LastIndex(line, ":")
			if f == 0 {
				go ping(line[f+13 : l])
			}
		}

		if line == "-------------------------" {
			if dmgstatus {
				dmgstatus = false

				if multiplayer {
					writefile(Config.Path, Config.File, dmgs)
				} else {
					if lastplayer != "" {
						writefile(Config.Path, Config.File, lastplayer)
					}
				}
				multiplayer = false
				lastplayer = ""
				dmgs = ""

			} else {
				dmgstatus = true
			}
		}

		if strings.Contains(line, "\"mp_c4timer\" = \"") { // Setzt den C4 Timer neu
			f := strings.Index(line, "\"mp_c4timer\" = \"")
			l := strings.Index(line[f+16:], "\"")
			if Config.Integration.Enable {
				i, err := strconv.Atoi(line[f+16 : l+16])
				if err != nil {
					panic(err)
				}
				GameStateIntegration.C4time = i
			}
		}

		if line == "SignalXWriteOpportunity(3)" { // New Round
			fmt.Println(" ")
			//go writedmg("")
		}

		if dmgstatus {
			if strings.Contains(line, "Damage Given to ") && strings.Contains(line, " in ") && strings.Contains(line, " hit") {
				fmt.Println("\x1b[32;1m", line, "\x1b[0m")
				go writedmg(line)
			}
		} else {
			if strings.Contains(line, "Damage Taken from ") && strings.Contains(line, " in ") && strings.Contains(line, " hit") {
				fmt.Println("\x1b[31;1m", line, "\x1b[0m")
			}
		}

		if line == "#end" {
			fmt.Print("\a")
			SmurfCheckerStatus = false
			go SmurfChecker.Start(Array, Config.APIKEY)
		}
		if SmurfCheckerStatus {
			Array = append(Array, line)
		}
		if strings.Index(line, "# userid name uniqueid connected ping loss state rate") == 0 {
			Array = []string{} // Array leeren
			SmurfCheckerStatus = true
		}

		// log.Printf("DEBUG: ",line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
}

var timerempty *time.Timer

func writefile(path, file, val string) {
	if val == "" {
		err := ioutil.WriteFile(path+"\\cfg\\"+file, []byte("echo byspddl"), 0644)
		if err != nil {
			panic(err)
		}
		fileIsEmpty = true
	} else {

		if timerempty != nil {
			timerempty.Stop()
		}

		// log.Print("say_team: ",val)
		err := ioutil.WriteFile(path+"\\cfg\\"+file, []byte("say_team "+val), 0644)
		if err != nil {
			panic(err)
		}
		fileIsEmpty = false

		lastplayer = ""
		dmgs = ""

		if timerempty != nil {
			timerempty.Stop()
		}
		timerempty = time.AfterFunc(30*time.Second, func() {
			if fileIsEmpty != true {
				err := ioutil.WriteFile(path+"\\cfg\\"+file, []byte("echo byspddl"), 0644)
				if err != nil {
					panic(err)
				}
			}
			fileIsEmpty = true
		})
	}

}

func writedmg(val string) {
	// Parse die HP
	f := strings.Index(val, " - ")
	l := strings.LastIndex(val, " in ")
	hp, _ := strconv.Atoi(val[f+3 : l])

	// TODO wenn 2 gegner über 100 hp verloren haben und es keinen anderen gibt sollte die Datei leeer sein
	if hp <= 99 { // Nur wenn der Gegner noch nicht Tot ist
		//log.Print("writedmg: ",val," (",hp,")")

		if lastplayer != "" {
			dmgs = dmgs + " - " + parsedmg(val)
			multiplayer = true
		} else {
			multiplayer = false
			dmgs = parsedmg(val)
		}
		lastplayer = val
	}
}

func parsedmg(line string) string {
	_firstPlayer := strings.Index(line, "\"")
	_lastPlayer := strings.LastIndex(line, "\"")
	player := line[_firstPlayer+1 : _lastPlayer]

	_firstDmg := strings.Index(line, " - ")
	dmg := line[_firstDmg+3:]
	return player + " " + dmg
}

func ping(ip string) {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		// fmt.Println(err)
		return
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
