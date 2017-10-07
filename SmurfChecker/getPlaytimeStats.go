package SmurfChecker

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type GetOwnedGames struct {
	Response struct {
		GameCount int `json:"game_count"`
		Games     []struct {
			Appid           int `json:"appid"`
			PlaytimeForever int `json:"playtime_forever"`
			Playtime2Weeks  int `json:"playtime_2weeks,omitempty"`
		} `json:"games"`
	} `json:"response"`
}

type Playtime struct {
	gameCount   int
	csgoForever int
	csgo2Week   int
}

func getPlaytimeStats(steamID64, APIKEY string) Playtime {
	// For public profiles only, this gets the following statistics and adds them as variables to self
	// number of games owned, playtime in csgo all time, playtime in csgo past 2 weeks, it also creates
	// the variable accountAge to self which contains the age of the account in years

	url := "http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=" + APIKEY + "&steamid=" + steamID64

	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	GetOwnedGames1 := GetOwnedGames{}
	jsonErr := json.Unmarshal(body, &GetOwnedGames1)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}

	gameCount := GetOwnedGames1.Response.GameCount

	csgoForever := 0
	csgo2Week := 0

	for _, e := range GetOwnedGames1.Response.Games {
		if e.Appid == 730 {
			csgoForever = e.PlaytimeForever / 60
			csgo2Week = e.Playtime2Weeks
		}
	}
	return Playtime{gameCount, csgoForever, csgo2Week}
}
