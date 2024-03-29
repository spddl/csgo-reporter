package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
)

type printTuple struct {
	steamID64      string
	csgoForever    int
	accountAge     int64
	accountCreated string
	gameCount      int
}

var APIKEY string

func Start(playerlist []string, apikey string) {
	steamID64List := make([]string, 0, len(playerlist))
	for _, line := range playerlist {
		r := regexp.MustCompile("#\\s+\\d+\\s\\d+\\s\"(.+?)\"\\s(.+? )")
		matches := r.FindAllStringSubmatch(line, -1)
		if len(matches) != 0 {
			steamID64List = append(steamID64List, strconv.Itoa(SteamidTo64bit(matches[0][2])))
		}
	}

	usersList := getPlayerSummaries(steamID64List, apikey)

	printTupleList := []printTuple{}
	for i := 0; i < len(usersList); i++ {
		a := usersList[i]
		if a.public {
			printTupleList = append(printTupleList, printTuple{
				steamID64:      a.name,
				csgoForever:    a.PlaytimeList.csgoForever,
				accountAge:     a.accountAge,
				accountCreated: a.accountCreated,
				gameCount:      a.PlaytimeList.gameCount,
			})
		} else {
			printTupleList = append(printTupleList, printTuple{
				steamID64:      "\x1b[31m" + a.name + "\x1b[37m",
				csgoForever:    0,
				accountAge:     0,
				accountCreated: "",
				gameCount:      0,
			})
		}
	}

	sort.Slice(printTupleList, func(i, j int) bool {
		return printTupleList[i].csgoForever < printTupleList[j].csgoForever
	})

	fmt.Printf("\n%11v | %11v | %5v | %8v | %v\n", "Acc created", "Acc Age(YR)", "Games", "CSGO HRS", "NAME")
	for _, u := range printTupleList {
		fmt.Printf("%11v | %11v | %5v | %8v | %v\n", u.accountCreated, strconv.Itoa(int(u.accountAge)), u.gameCount, strconv.Itoa(u.csgoForever), u.steamID64)
	}
	fmt.Println("")
	checkFriends(usersList)
}
