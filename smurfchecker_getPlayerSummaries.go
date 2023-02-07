package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
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
	steamID64      string
	name           string
	public         bool
	profileUrl     string
	accountCreated string
	accountAge     int64
	FriendsList    []Friends
	PlaytimeList   Playtime
}

func getPlayerSummaries(steamID64List []string, apikey string) []user {
	// Takes a list of steamID64s and returns a list of User objects containing the following information
	// username, steamid64, profileurl, isprofileprivate, and if the profile is not private it also adds the
	// date the account was created in unix timestamp

	if len(apikey) == 0 || strings.Contains(apikey, "#") {
		fmt.Println("\x1b[31;1mNO APIKEY\x1b[0m")
		fmt.Println("go to https://steamcommunity.com/dev/apikey")
		return []user{}
	}

	// create comma separated list of steamID64s
	url := "https://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=" + apikey + "&steamIDS=" + strings.Join(steamID64List, ",")

	client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Println(err)
		return []user{}
	}

	res, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return []user{}
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
		return []user{}
	}

	steamresp1 := GetPlayerSummaries{}
	err = json.Unmarshal(body, &steamresp1)
	if err != nil {
		log.Println(err)
		return []user{}
	}

	returnUserList := make([]user, 0, len(steamresp1.Response.Players))

	for i := 0; i < len(steamresp1.Response.Players); i++ {
		e := steamresp1.Response.Players[i]
		name := e.Personaname
		steamID64 := e.Steamid
		profileurl := e.Profileurl

		if e.Communityvisibilitystate == 3 { // public profile
			tm := time.Unix(int64(e.Timecreated), 0)
			returnUserList = append(returnUserList, user{
				steamID64:      steamID64,
				name:           name,
				public:         true,
				profileUrl:     profileurl,
				accountCreated: tm.Format("02.01.2006"),
				accountAge:     (time.Now().Unix() - int64(e.Timecreated)) / (60 * 60 * 24 * 365 /* One Year */),
				FriendsList:    getFriends(steamID64, apikey),
				PlaytimeList:   getPlaytimeStats(steamID64, apikey),
			})
		} else { // private profile
			returnUserList = append(returnUserList, user{
				steamID64:      steamID64,
				name:           name,
				public:         false,
				profileUrl:     profileurl,
				accountCreated: "",
				accountAge:     0,
				FriendsList:    []Friends{},
				PlaytimeList:   Playtime{},
			})
		}
	}
	return returnUserList
}
