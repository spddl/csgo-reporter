package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"

	"log"
	"os"
	"strings"

	"golang.org/x/sys/windows/registry"
)

type Config struct {
	Path           string `json:"path"`
	File           string `json:"file"`
	APIKEY         string `json:"apikey"`
	ReportTemplate string `json:"reportTemplate"`
	C4peep         string `json:"c4peep"`
	Integration    struct {
		Enable bool `json:"enable"`
		C4     bool `json:"c4"`
	} `json:"integration"`
	Sounds struct {
		C4 []struct {
			Time int64    `json:"time"`
			Mp3  []string `json:"mp3"`
		} `json:"c4"`
		NewRound string `json:"newRound"`
	} `json:"sounds"`
	Help bool `json:"help"`
}

const defaultConfigPath = "./config.json"

func (c *Config) Init() {
	var defaultPath string = pathJoin("C:", "Steam", "steamapps", "common", "Counter-Strike Global Offensive", "csgo")
	var defaultFile string = "report.cfg"

	flagPath := flag.String("path", defaultPath, "CSGO Folder")
	flagFile := flag.String("file", defaultFile, "cfg File")
	flag.Parse()

	if len(os.Args) < 2 { // keine Parameter
		if _, err := os.Stat(defaultConfigPath); os.IsNotExist(err) { // Keine config.json gefunden
			log.Println("Keine config.json gefunden")
			k, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\\Valve\\Steam`, registry.READ)
			if err != nil {
				fmt.Println(err)
			}

			s, _, err := k.GetStringValue("SteamPath")
			if err == nil {
				s = strings.ReplaceAll(s, "/", "\\")
				fmt.Printf("\x1b[32;1mFound Steam Folder:\x1b[0m%s\n\n", s)
				defaultPath = pathJoin(s, "steamapps", "common", "Counter-Strike Global Offensive", "csgo")
			}
			k.Close()

			for {
				*flagPath = readCLI("CSGO Path\ndefault: '" + defaultPath + "'\n")
				*flagPath = defaultPath

				b, err := ExistsFolder(*flagPath)
				if err != nil {
					log.Println(err)
				}
				if b {
					break
				} else {
					fmt.Printf("Ordner nicht gefunden\n\n")
				}
			}

			*flagFile = readCLI("Report File\ndefault: 'report.cfg'\n")
			if *flagFile == "" {
				*flagFile = "report.cfg"
			}

			c.File = *flagFile
			c.Path = *flagPath
			c.SaveFile()
		} else {
			c.LoadFile()
		}

	} else { // Parameter gefunden
		c.File = *flagFile
		c.Path = *flagPath
	}
}

func readCLI(txt string) string {
	reader := bufio.NewReader(os.Stdin)
	log.Println(txt)

	defaultPathbyte, _, err := reader.ReadLine()
	if err != nil {
		panic(err)
	}
	return string(defaultPathbyte)
}

func (c *Config) LoadFile() {
	file, e := os.ReadFile(defaultConfigPath)
	if e != nil {
		log.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	err := json.Unmarshal(file, &c)
	if err != nil {
		fmt.Println("error:", err)
	}
}

func (c *Config) SaveFile() {
	_, err := json.Marshal(c)
	if err != nil {
		panic("Config konnte nicht gespeichert werden (json error)")
	}
}
