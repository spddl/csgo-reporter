package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto"
	"github.com/nicememe/go-csgsi"
)

var ctScore int = 0
var tScore int = 0
var C4time int64 = 40
var otoCtx *oto.Context

type c4NewTimer struct {
	ticker    *time.Ticker
	isPlanted bool
	timers    []*time.Timer
	mu        sync.Mutex
	ctx       context.Context
	cancel    context.CancelFunc
}

func (c4 *c4NewTimer) planted(config *Config) {
	c4.mu.Lock()
	c4.ticker = time.NewTicker(time.Second)
	c4.isPlanted = true
	c4.timers = make([]*time.Timer, 0, len(config.Sounds.C4))

	c4.mu.Unlock()
	for _, value := range config.Sounds.C4 {
		go func(c4 *c4NewTimer, delay int64, mp3s []string) {
			if delay < 0 {
				return
			}
			t := time.AfterFunc(time.Duration(delay)*time.Second, func() {
				for _, mp3 := range mp3s {
					err := playSound(c4.ctx, otoCtx, mp3)
					if err != nil {
						log.Println(err)
					}
				}
			})
			c4.mu.Lock()
			c4.timers = append(c4.timers, t)
			c4.mu.Unlock()
		}(c4, C4time-value.Time, value.Mp3)
	}

	go func(tempTimer int64) {
		for {
			select {
			case <-c4.ticker.C:
				tempTimer -= 1
				if tempTimer < 1 {
					c4.cancel()
				}

				if tempTimer%5 == 0 || tempTimer < 11 {
					fmt.Println(tempTimer, "sek")
				}

			case <-c4.ctx.Done():
				c4.mu.Lock()
				c4.isPlanted = false
				c4.ticker.Stop()
				for _, timers := range c4.timers {
					timers.Stop()
				}
				c4.mu.Unlock()
				return
			}
		}
	}(C4time)
}

func GamestateIntegration(config *Config) {
	if !fileExists(pathJoin(config.Path, "cfg", "gamestate_integration_spddl.cfg")) {
		createGamestateIntegrationTemplate(config.Path)
		if err := exec.Command("powershell.exe", "Get-Process", "CSGO").Run(); err == nil {
			fmt.Println("\x1b[31;1mcsgo.exe is running, need CSGO Restart\x1b[0m")
		}
	}

	if len(config.Sounds.C4) != 0 || config.Sounds.NewRound != "" {
		c, err := oto.NewContext(48000, 2, 2, 8192)
		if err != nil {
			log.Println("err:", err)
		}
		otoCtx = c
	}

	game := csgsi.New(10)
	go func(game *csgsi.Game) {
		c4 := new(c4NewTimer)
		for state := range game.Channel {
			if !c4.isPlanted && state.Round != nil && state.Round.Bomb == "planted" && state.Round.Phase == "live" {

				ctx, cancel := context.WithCancel(context.Background())
				c4.ctx = ctx
				c4.cancel = cancel

				fmt.Println("\x1b[31;1mbomb has been planted\x1b[0m")
				go c4.planted(config)

			} else if c4.isPlanted && state.Round != nil && (state.Round.Phase == "over" || state.Round.Phase == "freezetime") {

				if state.Round.Bomb == "exploded" {
					fmt.Println("\x1b[31;1mbomb exploded\x1b[0m")
				} else if state.Round.Bomb == "defused" {
					fmt.Println("\x1b[36;1mbomb defused\x1b[0m")
				}

				c4.cancel()
			}

			if state.Round != nil && state.Map != nil && (state.Round.Phase == "over" || state.Round.Phase == "freezetime") {
				if ctScore != state.Map.Team_ct.Score || tScore != state.Map.Team_t.Score {
					fmt.Printf("%s --- \x1b[36;1m CT:%d\x1b[0m -\x1b[31;1m Ter:%d\x1b[0m ---\n\n", time.Now().Format("15:04:05"), state.Map.Team_ct.Score, state.Map.Team_t.Score)
					ctScore = state.Map.Team_ct.Score
					tScore = state.Map.Team_t.Score

					if config.Sounds.NewRound != "" {
						ctx, cancel := context.WithCancel(context.Background())
						err := playSound(ctx, otoCtx, config.Sounds.NewRound)
						if err != nil {
							log.Println(err)
						}
						cancel()
					}

				}
			}
		}
	}(game)
	err := game.Listen(":1337")
	if err != nil {
		fmt.Println(err)
	}
}

func createGamestateIntegrationTemplate(csgoPath string) {
	gamestate_integration_file := pathJoin(csgoPath, "cfg", "gamestate_integration_spddl.cfg")

	var file, err = os.OpenFile(gamestate_integration_file, os.O_RDONLY|os.O_CREATE, 0666)
	_, err = file.WriteString(`"csgo-reporter"{"uri" "http://127.0.0.1:1337" "timeout" "5.0" "buffer" "0.1" "throttle" "0.5" "heathbeat" "60.0" "data"{"map" "1" "round" "1"}}`)
	if err != nil {
		log.Println(err)
	}
	file.Close()
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func playSound(ctx context.Context, c *oto.Context, filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	d, err := mp3.NewDecoder(f)
	if err != nil {
		return err
	}
	p := c.NewPlayer()
	_ = copy(ctx, p, d)

	p.Close()
	f.Close()
	return nil
}
