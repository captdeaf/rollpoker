package rollpoker

import (
	"time"
	"fmt"
	"math/rand"
	"strings"
	"strconv"
)

func (player *Player) DoStartGame(game *Game, gc *GameCommand) bool {
	// Sanity checks:
	if game.Public.State != NOGAME { return false }
	if len(game.Public.Players) < 2 { return false }
	// TODO: MTT support
	if len(game.Public.Players) > 10 { return false }

	// So, we're trying to start a game.
	// 1) Do we have enough players? Do we need multiple tables?
	allPlayers := []string{}

	// Deep copy the Settings structure.
	settings := *(game.Private.OrigState)
	game.Public.GameSettings = &settings

	for _, player := range game.Public.Players {
		allPlayers = append(allPlayers, player.PlayerId)
		player.Chips = settings.StartingChips
		player.Bet = 0
		player.Rank = 0
		player.State = ""
		player.Hand = []string{}
	}
	if len(allPlayers) > 10 {
		// TODO: Multiple tables
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
	table.Dealer = ""
	game.Public.Tables["table0"] = &table
	table.Dolist = GAME_COMMANDS["texasholdem"]

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

	// TODO: Trigger start on all tables

	// Don't start the RunCommands - this is special, we want to
	// run the start for All tables by hand
	i := time.Duration(1)
	for tname, _ := range game.Public.Tables {
		fmt.Printf("Calling RunCommands for %dth time\n", i)
		go RunCommands(game.Name, tname, i)
		i += 1
	}
	return false
}

func (player *Player) DoBet(game *Game, gc *GameCommand) bool {
	if player.State != TURN { return false }
	tablename := game.TableForPlayer(player)
	ibet, _ := strconv.ParseInt(gc.Args["amount"], 10, 32)
	ibet = 50 // Temp override
	// TODO: Ensure they can bet. Minimum is blind *(or last raise),
	// Maximum is their amount of chips. (if chips < min, min = chips)
	fmt.Printf("Player %s bets %d", player.DisplayName, ibet)
	DoBet(game, tablename, gc.PlayerId, int(ibet), false)
	return true
}
