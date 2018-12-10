package mahjong

// GetPlayerList returns the list of player's name
func (room Room) GetPlayerList() []string {
	var nameList []string
	for _, player := range room.Players {
		nameList = append(nameList, player.Name())
	}
	return nameList
}

// GetReadyPlayers returns the name list of ready player
func (room Room) GetReadyPlayers() []string {
	var nameList []string
	for _, player := range room.Players {
		nameList = append(nameList, player.Name())
	}
	return nameList
}

// GetLack returns each player's lack
func (room Room) GetLack() []int {
	if room.State < ChooseLack {
		return []int{}
	}
	var lacks []int
	for _, player := range room.Players {
		lacks = append(lacks, player.Lack)
	}
	return lacks
}

// GetHandCount returns each player's amount of hand
func (room Room) GetHandCount() []int {
	if room.State < IdxTurn {
		return []int{}
	}
	var amounts []int
	for _, player := range room.Players {
		amounts = append(amounts, int(player.Hand.Count()))
	}
	return amounts
}

// GetRemainCount returns amount of deck
func (room Room) GetRemainCount() int {
	if room.State < IdxTurn {
		return 56
	}
	return int(room.Deck.Count())
}

// GetDoor returns each player's door
func (room Room) GetDoor(id int) ([][]string, []int, bool) {
	if room.State < IdxTurn {
		return [][]string{}, []int{}, true
	}
	var inVisibleList []int
	var visibleList   [][]string
	for _, player := range room.Players {
		if id == player.ID {
			visibleList   = append(visibleList, player.Door.ToStringArray())
			inVisibleList = append(inVisibleList, 0)
		} else {
			visibleList   = append(visibleList, player.VisiableDoor.ToStringArray())
			inVisibleList = append(inVisibleList, int(player.Door.Count() - player.VisiableDoor.Count()))
		}
	}
	return visibleList, inVisibleList, false
}

// GetSea returns each player's discard card
func (room Room) GetSea() ([][]string, bool) {
	if room.State < IdxTurn {
		return [][]string{}, true
	}
	var discardTileList [][]string
	for _, player := range room.Players {
		discardTileList = append(discardTileList, player.DiscardTiles.ToStringArray())
	}
	return discardTileList, false
}

// GetHu returns each player's hu card
func (room Room) GetHu() ([][]string, bool) {
	if room.State < IdxTurn {
		return [][]string{}, true
	}
	var huList [][]string
	for _, player := range room.Players {
		huList = append(huList, player.HuTiles.ToStringArray())
	}
	return huList, false
}

// GetCurrentIdx returns current index
func (room Room) GetCurrentIdx() int {
	id := -1
	if room.State >= IdxTurn {
		id = room.State - IdxTurn
	}
	return id
}

// GetScore returns each player's score
func (room Room) GetScore() []int {
	var scoreList []int
	for _, player := range room.Players {
		scoreList = append(scoreList, player.Credit)
	}
	return scoreList
}
