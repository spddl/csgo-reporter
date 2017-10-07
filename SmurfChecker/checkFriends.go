package SmurfChecker

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type CFriends struct {
	userPersonaName      string
	checkUserPersonaName string
	timeFriends          int64
}

func contains(s []CFriends, c, e user) bool {
	for _, v := range s {
		if v.checkUserPersonaName == c.name && v.userPersonaName == e.name {
			return true
		}
		if v.checkUserPersonaName == e.name && v.userPersonaName == c.name {
			return true
		}
	}
	return false
}

func checkFriends(usersList []user) {
	// Checks for users inside usersList that are friends with each other, it loops through every user, then for every
	// user loops through every friend, then for every friend it loops through every user again, it checks if the
	// friend has the same steamID64 as the user
	// Returns a list of tuples that contain (username1, username2, howLongInDaysBeenFriends)

	currentFriends := []CFriends{}
	for _, user := range usersList {
		if user.public == false {
			continue
		}

		for _, f := range user.FriendsList {
			for _, checkuser := range usersList {
				if checkuser.steamID64 == f.Steamid {

					if !contains(currentFriends, checkuser, user) { // TODO: wenn 3 leute befreundet sind gibs probleme
						timeFriends := (time.Now().Unix() - int64(f.FriendSince)) / (60 * 60 * 24)
						currentFriends = append(currentFriends, CFriends{user.name, checkuser.name, timeFriends})
					}

				}
			}
		}

	}

	sort.Slice(currentFriends, func(i, j int) bool {
		return currentFriends[i].timeFriends < currentFriends[j].timeFriends
	})

	for _, e := range currentFriends {
		fmt.Println("\x1b[36m\""+strings.Trim(e.checkUserPersonaName, " ")+"\" and \""+strings.Trim(e.userPersonaName, " ")+"\" have been friends for", e.timeFriends, "days\x1b[0m")
	}
}
