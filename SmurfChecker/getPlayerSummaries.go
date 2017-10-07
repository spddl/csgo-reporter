package SmurfChecker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type GetPlayerSummaries struct {
	Response struct {
		Players []struct {
			Steamid                  string `json:"steamid"`
			Communityvisibilitystate int    `json:"communityvisibilitystate"`
			Profilestate             int    `json:"profilestate"`
			Personaname              string `json:"personaname"`
			Lastlogoff               int    `json:"lastlogoff"`
			Commentpermission        int    `json:"commentpermission"`
			Profileurl               string `json:"profileurl"`
			Avatar                   string `json:"avatar"`
			Avatarmedium             string `json:"avatarmedium"`
			Avatarfull               string `json:"avatarfull"`
			Personastate             int    `json:"personastate"`
			Realname                 string `json:"realname"`
			Primaryclanid            string `json:"primaryclanid"`
			Timecreated              int    `json:"timecreated"`
			Personastateflags        int    `json:"personastateflags"`
			Loccountrycode           string `json:"loccountrycode"`
		} `json:"players"`
	} `json:"response"`
}

type user struct {
	steamID64    string
	name         string
	public       bool
	profileUrl   string
	timeCreated  string
	accountAge   int64
	FriendsList  []Friends
	PlaytimeList Playtime
}

func getPlayerSummaries(Array []string, APIKEY string) []user {
	// Takes a list of steamID64s and returns a list of User objects containing the following information
	// username, steamid64, profileurl, isprofileprivate, and if the profile is not private it also adds the
	// date the account was created in unix timestamp

	if len(APIKEY) == 0 || strings.Contains(APIKEY, "#") {
		fmt.Println("\x1b[31;1mNO APIKEY\x1b[0m")
		fmt.Println("go to https://steamcommunity.com/dev/apikey")
		return []user{}
	}

	// create comma separated list of steamID64s
	url := "http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=" + APIKEY + "&steamIDS=" + strings.Join(Array[:], ",")

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
		log.Fatal("read Err ", readErr)
	}

	steamresp1 := GetPlayerSummaries{}
	jsonErr := json.Unmarshal(body, &steamresp1)
	if jsonErr != nil {
		log.Fatal("json Err ", jsonErr)
	}

	returnUserList := []user{}
	for _, e := range steamresp1.Response.Players {
		name := e.Personaname
		steamID64 := e.Steamid
		profileurl := e.Profileurl

		if e.Communityvisibilitystate == 3 { // public profile
			timeCreated := e.Timecreated

			returnUserList = append(returnUserList, user{steamID64, name, true, profileurl, strconv.Itoa(timeCreated), (time.Now().Unix() - int64(timeCreated)) / (60 * 60 * 24 * 365), getFriends(steamID64, APIKEY), getPlaytimeStats(steamID64, APIKEY)})
		} else { // private profile
			returnUserList = append(returnUserList, user{steamID64, name, false, profileurl, "", 0, []Friends{}, Playtime{}}) // TODO evtl. auch ohne ""
		}
	}
	return returnUserList
}
