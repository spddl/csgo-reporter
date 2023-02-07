package main

import (
	"fmt"
	"strings"
	"time"
)

type CFriends struct {
	userPersonaName      string
	checkUserPersonaName string
	timeFriends          int64
}

func checkFriends(usersList []user) {
	// Checks for users inside usersList that are friends with each other, it loops through every user, then for every
	// user loops through every friend, then for every friend it loops through every user again, it checks if the
	// friend has the same steamID64 as the user
	// Returns a list of tuples that contain (username1, username2, howLongInDaysBeenFriends)

	for i := 0; i < len(usersList); i++ {
		user := usersList[i]

		if !user.public { // Profil ist Privat
			continue
		}

		currentFriends := []CFriends{}
		for _, f := range user.FriendsList {

			for i := 0; i < len(usersList); i++ {
				StatusPlayerList := usersList[i]

				if StatusPlayerList.steamID64 == f.Steamid {
					currentFriends = append(currentFriends, CFriends{
						userPersonaName:      user.name,
						checkUserPersonaName: StatusPlayerList.name,
						timeFriends:          (time.Now().Unix() - int64(f.FriendSince)) / (60 * 60 * 24),
					})
					break
				}
			}
		}

		if len(currentFriends) != 0 {
			var tempArray []string
			for _, ff := range currentFriends {
				tempArray = append(tempArray, fmt.Sprintf("\x1b[36m%s for %d days\x1b[0m", ff.checkUserPersonaName, ff.timeFriends))
			}
			if len(currentFriends) == 1 {
				fmt.Printf("\"%s\" have been friends: %v\n", user.name, strings.Join(tempArray, " and "))
			} else {
				fmt.Printf("\"%s\" have been friends(%d): %v\n", user.name, len(currentFriends), strings.Join(tempArray, " and "))
			}
		}
	}
}
