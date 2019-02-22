package mahjong

// GetPlayerList returns the list of player's name
func (room Room) GetPlayerList() []string {
	var nameList []string
	for _, player := range room.Players {
		nameList = append(nameList, player.Name())
	}
	return nameList
}

// GetWindAndRound returns the wind and round
func (room Room) GetWindAndRound() (int, int) {
	return room.Wind, room.Round
}

// GetReadyPlayers returns the name list of ready player
func (room Room) GetReadyPlayers() []string {
	var nameList []string
	for _, player := range room.Players {
		nameList = append(nameList, player.Name())
	}
	return nameList
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
func (room Room) GetDoor(id int) ([][][]string, []int, bool) {
	if room.State < IdxTurn {
		return [][][]string{}, []int{}, true
	}
	var inVisibleList []int
	var visibleList   [][][]string
	for _, player := range room.Players {
		if id == player.ID {
			visibleList   = append(visibleList, [][]string{ player.EatTiles.ToStringArray(), player.PonTiles.ToStringArray(), player.GonTiles.ToStringArray(), player.OngonTiles.ToStringArray() })
			inVisibleList = append(inVisibleList, 0)
		} else {
			visibleList   = append(visibleList, [][]string{ player.EatTiles.ToStringArray(), player.PonTiles.ToStringArray(), player.GonTiles.ToStringArray(), []string{} })
			inVisibleList = append(inVisibleList, int(player.OngonTiles.Count()))
		}
	}
	return visibleList, inVisibleList, false
}

// GetSea returns each player's discard tile
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

// GetFlower returns each player's flower tile
func (room Room) GetFlower() ([][]string, bool) {
	var flowerList [][]string
	for _, player := range room.Players {
		flowerList = append(flowerList, player.Flowers.ToStringArray())
	}
	return flowerList, false
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
