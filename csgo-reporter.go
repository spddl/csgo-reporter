package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"path"
	"strconv"
	"strings"
	"time"

	"./SmurfChecker"

	"github.com/tatsushid/go-fastping"
)

var fileIsEmpty = true
var SmurfCheckerStatus = false

var Playerlist []string

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.Lmicroseconds)

	dmg := new(dmgParser)

	config := new(Config)
	config.Init()

	if config.Help {
		fmt.Println("\x1b[36;1madd Steam Parameter: -condebug +bind \",\" report.cfg\x1b[0m")
	}

	if config.Integration.Enable {
		go GamestateIntegration(config) // https://github.com/dank/go-csgsi
	}

	// Leeren der console.log
	err := ioutil.WriteFile(path.Join(config.Path, "console.log"), []byte{10}, 0644)
	if err != nil {
		log.Println(err)
	}

	// Leeren der report.cfg
	WriteFile(config, []byte{})

	t, err := newTailReader(path.Join(config.Path, "console.log"))
	if err != nil {
		log.Fatal(err)
	}
	defer t.Close()
	scanner := bufio.NewScanner(t)
	for scanner.Scan() {
		line := scanner.Text()

		switch true {
		case strings.Contains(line, "Confirmed best official datacenter ping: "): // errechnete Verzögung von Valve:
			fmt.Printf("\x1b[36;1m%s\x1b[0m\n", line)

		case strings.HasPrefix(line, "udp/ip  : "):
			go ping(line[10:strings.LastIndex(line, ":")])
		case strings.HasPrefix(line, "Connected to "):
			if strings.Count(line, ":") == 1 { // zuviel ist igendeine ID und zuwenig könnte "loopback" sein
				go ping(line[13:strings.LastIndex(line, ":")])
			}

		case line == "-------------------------":
			if dmg.Enable {
				dmg.Output(config)
			}
			dmg.Enable = !dmg.Enable

		case dmg.Enable && strings.Contains(line, "Damage Given to ") && strings.Contains(line, " in ") && strings.Contains(line, " hit"):
			fmt.Println("\x1b[32;1m", line, "\x1b[0m")
			dmg.Add(line)

		case !dmg.Enable && strings.Contains(line, "Damage Taken from ") && strings.Contains(line, " in ") && strings.Contains(line, " hit"):
			fmt.Println("\x1b[31;1m", line, "\x1b[0m")

		case strings.HasPrefix(line, `"mp_c4timer" = "`):
			if config.Integration.Enable {
				l := strings.Index(line[16:], `"`)
				value := line[16 : l+16]
				if strings.Contains(value, ".") { // der Timer ist in int und in csgo vermutlich ein float
					value = value[:strings.Index(value, ".")]
				}
				c4time, err := strconv.Atoi(value) // TODO string to int64 ?
				if err != nil {
					panic(err)
				}
				C4time = int64(c4time)
			}

		case line == "SignalXWriteOpportunity(3)": // New Round
			WriteFile(config, []byte{})

		case strings.Index(line, "# userid name uniqueid connected ping loss state rate") == 0: // beginn from "status"
			Playerlist = []string{} // Array leeren
			SmurfCheckerStatus = true

		case line == "#end":
			SmurfCheckerStatus = false
			go SmurfChecker.Start(Playerlist, config.APIKEY)

		default:
			if SmurfCheckerStatus { // add every player
				Playerlist = append(Playerlist, line)
			}
			// log.Println("DEBUG: ", line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

var timerempty *time.Timer

var echoBySpddl = []byte{101, 99, 104, 111, 32, 98, 121, 115, 112, 100, 100, 108} // echo byspddl
var say_team = []byte{115, 97, 121, 95, 116, 101, 97, 109, 32, 34}                // say_team "

func WriteFile(c *Config, val []byte) {
	var reportPath = path.Join(c.Path, "cfg", c.File)

	if bytes.Equal(val, []byte{}) {
		err := ioutil.WriteFile(reportPath, echoBySpddl, 0644)
		if err != nil {
			log.Println(err)
		}
		fileIsEmpty = true
	} else {
		err := ioutil.WriteFile(reportPath, append(say_team, append(val, []byte{34}...)...), 0644)
		if err != nil {
			log.Println(err)
		}
		fileIsEmpty = false

		if timerempty != nil {
			timerempty.Stop()
		}
		timerempty = time.AfterFunc(2*time.Minute, func() {
			if !fileIsEmpty {
				err := ioutil.WriteFile(reportPath, echoBySpddl, 0644)
				if err != nil {
					log.Println(err)
				}
			}
			fileIsEmpty = true
		})
	}
}

func ping(ip string) {
	p := fastping.NewPinger()
	ra, err := net.ResolveIPAddr("ip4:icmp", ip)
	if err != nil {
		log.Println(err)
		return
	}
	p.AddIPAddr(ra)
	p.OnRecv = func(addr *net.IPAddr, rtt time.Duration) {
		fmt.Printf("\x1b[36;1mIP: %s, RTT: %v\x1b[0m\n", addr.String(), rtt) //  Round trip time https://de.wikipedia.org/wiki/Ping_(Daten%C3%BCbertragung)
	}
	err = p.Run()
	if err != nil {
		fmt.Println(err)
	}
}

// Powershell tail -f
// cmd := exec.Command("powershell.exe", "Get-Content", "'"+config.Path+"\\console.log'", "-Wait", "-Encoding UTF8")
// // cmd := exec.Command("powershell.exe", "Get-Content", "'"+Config.Path+"\\console.log'", "-Wait", "-Encoding Oem")
// stdout, err := cmd.StdoutPipe()
// if err != nil {
// 	log.Fatal(err)
// }
// if err := cmd.Start(); err != nil {
// 	log.Fatal(err)
// }

// scanner := bufio.NewScanner(stdout)
// for scanner.Scan() {
// 	line := scanner.Text()
// }

// if err := cmd.Wait(); err != nil {
// 	log.Fatal(err)
// }
