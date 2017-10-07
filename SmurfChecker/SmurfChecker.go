package SmurfChecker

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/xlab/tablewriter"
)

type printTuple struct {
	steamID64   string
	csgoForever int
	accountAge  int64
}

var APIKEY string

func Start(status []string, APIKEY string) {
	APIKEY = APIKEY
	steamID64List := []string{} // Array leeren
	for _, line := range status {
		r, e := regexp.Compile("#\\s+\\d+\\s\\d+\\s\"(.+?)\"\\s(.+? )")
		if e != nil {
			panic(e)
		}

		matches := r.FindAllStringSubmatch(line, -1)
		if len(matches) != 0 {
			steamID64List = append(steamID64List, strconv.Itoa(SteamidTo64bit(matches[0][2])))
		}
	}

	usersList := getPlayerSummaries(steamID64List, APIKEY)

	printTupleList := []printTuple{}
	for _, a := range usersList {
		if a.public {
			printTupleList = append(printTupleList, printTuple{a.name, a.PlaytimeList.csgoForever, a.accountAge})
		} else {
			printTupleList = append(printTupleList, printTuple{"\x1b[31m" + a.name + "\x1b[37m", 0, 0})
		}
	}

	sort.Slice(printTupleList, func(i, j int) bool {
		return printTupleList[i].csgoForever < printTupleList[j].csgoForever
	})

	table := tablewriter.CreateTable()

	table.AddHeaders("NAME", "HRS", "YR")
	for _, u := range printTupleList {
		table.AddRow(u.steamID64, strconv.Itoa(u.csgoForever), strconv.Itoa(int(u.accountAge)))
	}
	fmt.Println(table.Render())

	fmt.Println(" ")

	go checkFriends(usersList)
}
