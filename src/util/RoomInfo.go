package util

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
	var res []int
	for _, player := range room.Players {
		res = append(res, player.Lack)
	}
	return res
}

// GetHandCount returns each player's amount of hand
func (room Room) GetHandCount() []int {
	if room.State < IDTurn {
		return []int{}
	}
	var res []int
	for _, player := range room.Players {
		res = append(res, int(player.Hand.Count()))
	}
	return res
}

// GetRemainCount returns amount of deck
func (room Room) GetRemainCount() int {
	if room.State < IDTurn {
		return 56
	}
	return int(room.Deck.Count())
}

// GetDoor returns each player's door
func (room Room) GetDoor(id int) ([][]string, []int, bool) {
	if room.State < IDTurn {
		return [][]string{}, []int{}, true
	}
	var inVisible []int
	var res       [][]string
	for _, player := range room.Players {
		if id == player.ID {
			res       = append(res, player.Door.ToStringArray())
			inVisible = append(inVisible, 0)
		} else {
			res       = append(res, player.VisiableDoor.ToStringArray())
			inVisible = append(inVisible, int(player.Door.Count() - player.VisiableDoor.Count()))
		}
	}
	return res, inVisible, false
}

// GetSea returns each player's discard card
func (room Room) GetSea() ([][]string, bool) {
	if room.State < IDTurn {
		return [][]string{}, true
	}
	var res [][]string
	for _, player := range room.Players {
		res = append(res, player.DiscardTiles.ToStringArray())
	}
	return res, false
}

// GetHu returns each player's hu card
func (room Room) GetHu() ([][]string, bool) {
	if room.State < IDTurn {
		return [][]string{}, true
	}
	var res [][]string
	for _, player := range room.Players {
		res = append(res, player.HuTiles.ToStringArray())
	}
	return res, false
}

// GetCurrentIdx returns current index
func (room Room) GetCurrentIdx() int {
	id := -1
	if room.State >= IDTurn {
		id = room.State - IDTurn
	}
	return id
}

// GetScore returns each player's score
func (room Room) GetScore() []int {
	var res []int
	for _, player := range room.Players {
		res = append(res, player.Credit)
	}
	return res
}
