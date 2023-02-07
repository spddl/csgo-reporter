package main

import (
	"strconv"
	"strings"
)

func SteamidTo64bit(steamid string) int {
	// credit to MattDMo at http://stackoverflow.com/a/36463885/2056979
	// converts a steamid in format (STEAM_0:XXXXXXX) to a steamID64 (83294723984234)

	steam64id := 76561197960265728
	id_split := strings.Split(strings.Trim(steamid, " "), ":")

	i, err := strconv.Atoi(id_split[2])
	if err != nil {
		panic(err)
	}
	steam64id += i * 2

	if id_split[1] == "1" {
		steam64id += 1
	}

	return steam64id
}
