package GameStateIntegration

import (
  "fmt"
  "time"
  "log"
  "github.com/nicememe/go-csgsi"
  "math"
)

var isPlanted bool = false
var ctScore int = 0
var tScore int = 0


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


func Start() {
  game := csgsi.New(10)

	go func() {

		for state := range game.Channel {

			if !isPlanted && state.Round != nil && state.Round.Bomb == "planted" && state.Round.Phase == "live" {
        log.Print("\x1b[31;1mbomb has been planted\x1b[0m")

				go c4timer()
				isPlanted = true
			} else if isPlanted && state.Round != nil && (state.Round.Phase == "over" || state.Round.Phase == "freezetime") {

        if state.Round.Bomb == "exploded" {
          log.Print("\x1b[31;1mbomb exploded\x1b[0m")
        } else if state.Round.Bomb == "defused" {
          log.Print("\x1b[36;1mbomb defused\x1b[0m")
        }

				isPlanted = false
			}

			if (state.Round != nil && (state.Round.Phase == "over" || state.Round.Phase == "freezetime")) {
				if ctScore != state.Map.Team_ct.Score || tScore != state.Map.Team_t.Score {
          log.Print("--- \x1b[36;1m CT:",state.Map.Team_ct.Score,"\x1b[0m -\x1b[31;1m Ter:",state.Map.Team_t.Score,"\x1b[0m ---")
					ctScore = state.Map.Team_ct.Score
					tScore = state.Map.Team_t.Score
				}
			}


		}
	}()
	game.Listen(":3000")
}

func countdown(ticker Ticker, duration time.Duration) chan time.Duration {
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

// TODO "mp_c4timer" = "40" min. 10.000000 client notify replicated - how long from when the C4 is armed until it blows
func c4timer() {
  for d := range countdown(NewTicker(time.Second), 40*time.Second) {
    if isPlanted {
      num := float64(d)/1000000000
      if math.Mod(num, 5) == 0 || num < 5 {
        if num > 0 {
          fmt.Println(d)
        }
      }
    } else {
      break
    }
  }
}
