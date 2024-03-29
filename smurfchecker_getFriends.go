package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type GetFriendList struct {
	Friendslist struct {
		Friends []struct {
			Steamid      string `json:"steamid"`
			Relationship string `json:"relationship"`
			FriendSince  int    `json:"friend_since"`
		} `json:"friends"`
	} `json:"friendslist"`
}

type Friends struct {
	Steamid     string
	FriendSince int
}

func getFriends(steamID64, apikey string) []Friends {
	// Gets a list of tuples containing all the friends that a user has and the unix timestamp that is the date
	// that the two users became friends. Returns nothing, just adds it to self

	url := "https://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=" + apikey + "&steamid=" + steamID64 + "&relationship=friend"

	client := http.Client{
		Timeout: time.Second * 5, // Maximum of 5 secs
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Fatal(err)
	}

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}

	GetFriendList1 := GetFriendList{}
	jsonErr := json.Unmarshal(body, &GetFriendList1)
	if jsonErr != nil {
		fmt.Println(url)
		fmt.Println(string(body))
		log.Println(jsonErr)
	}

	FriendsList := make([]Friends, 0, len(GetFriendList1.Friendslist.Friends))
	for _, e := range GetFriendList1.Friendslist.Friends {
		FriendsList = append(FriendsList, Friends{e.Steamid, e.FriendSince})
	}

	return FriendsList
}
