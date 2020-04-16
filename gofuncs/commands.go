package rollpoker

import (
	"time"
	"math/rand"
	"strings"
	"strconv"
)

func (game *Game) StartGame(player *Player, gc *GameCommand) bool {
	// if game.Public.State != NOGAME {
		// return false
	// }

	// So, we're trying to start a game.
	// 1) Do we have enough players? Do we need multiple tables?
	allPlayers := []string{}
	settings := game.Private.OrigState
	for _, player := range game.Public.Players {
		allPlayers = append(allPlayers, player.PlayerId)
		player.Chips = settings.StartingChips
		player.Bet = 0
		player.Rank = 0
		player.State = ""
	}
	if len(allPlayers) < 2 {
		return false
	}
	if len(allPlayers) > 10 {
		// TODO: Multiple tables
		return false
	}
	// 2) Shuffle the players and assign them seats.
	// We assign basically bouncing back and forth across the table to
	// fill in seats spaced appropriately.
	allSeats := []string{
		"seat2", "seat7", "seat9", "seat4", "seat1",
		"seat6", "seat0", "seat5", "seat8", "seat3",
	}
	rand.Shuffle(len(allPlayers), func(i, j int) {
		allPlayers[i], allPlayers[j] = allPlayers[j], allPlayers[i]
	})

	// 3) Populate TableState from GameSettings and choose Dealer at random.
	game.Public.Tables = make(map[string]*TableState)
	table := TableState{}
	table.Seats = map[string]string{};
	for seatnum, pid := range allPlayers {
		table.Seats[allSeats[seatnum]] = pid
	}
	// First player is dealer, because player position is random, anyway.
	table.Dealer = allPlayers[0]
	game.Public.Tables["table0"] = &table

	game.Public.GameSettings = game.Private.OrigState

	blindstr := game.Public.GameSettings.BlindStructure[0]
	if len(game.Public.GameSettings.BlindStructure) > 0 {
		game.Public.GameSettings.BlindStructure = game.Public.GameSettings.BlindStructure[1:]
	}

	blindsplit := strings.Fields(blindstr)
	game.Public.CurrentBlinds = make([]int,len(blindsplit))
	game.Public.BlindTime = 0 // TODO: Now + blindTime.
	game.Public.PausedAt = 0
	game.Public.State = INGAME

	for i, val := range blindsplit {
		ival, _ := strconv.ParseInt(val, 10, 32)
		game.Public.CurrentBlinds[i] = int(ival)
	}

	SaveGame(game)

	go WaitAndDeal(game)

	return true
}

func WaitAndDeal(game *Game) {
	time.Sleep(5 * time.Second)

	// find Dealer, set up rotation, and deal.
}
