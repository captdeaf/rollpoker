// Player Commands for Poker
package rollpoker

import (
	"time"
	"fmt"
	"math/rand"
	"strings"
	"strconv"
)

func (player *Player) TryPokerKick(game *RoomData, gc *GameCommand) int {
	return SAVE|RUN
}

func (player *Player) TrySignupStartPoker(game *RoomData, gc *GameCommand) *CommandResponse {
	// Sanity checks:
	if len(game.Room.Players) < 2 { return CError("Not enough players") }
	// TODO: MTT support
	if len(game.Room.Players) > 10 { return CError("Too many players for now") }

	// So, we're trying to start a game.
	// 1) Do we have enough players? Do we need multiple tables?
	allPlayers := []string{}
	settings := game.Room.OrigSettings
	game.Room.GameSettings = settings

	for _, player := range game.Room.Players {
		allPlayers = append(allPlayers, player.PlayerId)
		player.Chips = settings.StartingChips
		player.Bet = 0
		player.Rank = 0
		player.State = ""
		player.Hand = make([]string, 0)
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
	game.Room.Tables = make(map[string]*TableState)
	table := TableState{}
	table.Seats = map[string]string{};
	for seatnum, pid := range allPlayers {
		table.Seats[allSeats[seatnum]] = pid
	}
	// First player is dealer, because player position is random, anyway.
	table.Dealer = ""
	game.Room.Tables["table0"] = &table
	table.Dolist = make(GameDef, len(GAME_COMMANDS["texasholdem"]))
	copy(table.Dolist, GAME_COMMANDS["texasholdem"])
	fmt.Printf("Got: %v\n", table.Dolist)

	blindstr := game.Room.GameSettings.BlindStructure[0]
	if len(game.Room.GameSettings.BlindStructure) > 0 {
		game.Room.GameSettings.BlindStructure = game.Room.GameSettings.BlindStructure[1:]
	}

	blindsplit := strings.Fields(blindstr)
	game.Room.CurrentBlinds = make([]int,len(blindsplit))
	game.Room.BlindTime = 0 // TODO: Now + blindTime.
	game.Room.PausedAt = 0
	game.Room.RoomState = POKER

	for i, val := range blindsplit {
		ival, _ := strconv.ParseInt(val, 10, 32)
		game.Room.CurrentBlinds[i] = int(ival)
	}

	// TODO: Trigger start on all tables

	// Don't start the RunCommands - this is special, we want to
	// run the start for All tables by hand
	i := time.Duration(1)
	for tname, _ := range game.Room.Tables {
		go RunCommands(game.Name, tname, i)
		i += 1
	}
	LogMessage(game, "%s starts the game", player.DisplayName)
	// We handle RunCommands, so we return special CommandResponse instead of COK
	return CResponse("", false, true, true)
}

func (player *Player) TryPokerCheck(game *RoomData, gc *GameCommand) int {
	if player.State != TURN { return ERR }
	tablename := game.TableForPlayer(player)
	table := game.Room.Tables[tablename]

	remaining := table.CurBet - player.Bet
	if remaining > 0 {
		return ERR
	}

	DoCall(game, tablename, gc.PlayerId, 0)
	LogEvent(game, "Call", player.PlayerId, remaining, "CHECK")
	LogMessage(game, "%s checks", player.DisplayName)
	return SAVE|RUN
}

func (player *Player) TryPokerFold(game *RoomData, gc *GameCommand) int {
	if player.State != TURN { return ERR }
	tablename := game.TableForPlayer(player)

	DoFold(game, tablename, gc.PlayerId)
	LogEvent(game, "Fold", player.PlayerId, "FOLD")
	LogMessage(game, "%s folds", player.DisplayName)
	return SAVE|RUN
}

func (player *Player) TryPokerCall(game *RoomData, gc *GameCommand) int {
	if player.State != TURN { return ERR }
	tablename := game.TableForPlayer(player)
	table := game.Room.Tables[tablename]

	remaining := table.CurBet - player.Bet
	if remaining > player.Chips {
		remaining = player.Chips
	}

	// Maximum is their amount of chips. (if chips < min, min = chips)
	DoCall(game, tablename, gc.PlayerId, remaining)
	if remaining == 0 {
		LogEvent(game, "Call", player.PlayerId, remaining, "CHECK")
		LogMessage(game, "%s checks", player.DisplayName)
	} else {
		LogEvent(game, "Call", player.PlayerId, remaining, "CALL")
		LogMessage(game, "%s calls for %d", player.DisplayName, remaining)
	}
	return SAVE|RUN
}

func (player *Player) TryPokerBet(game *RoomData, gc *GameCommand) int {
	// This can be a Check, a Call, a Bet, or a Raise, depending on amount.
	if player.State != TURN { return ERR }
	tablename := game.TableForPlayer(player)
	table := game.Room.Tables[tablename]
	i64bet, ierr := strconv.ParseInt(gc.Args["amount"], 10, 32)
	ibet := int(i64bet)

	if ierr != nil || ibet < 0 { return ERR }

	// Table current
	curbet := table.CurBet
	total := ibet + player.Bet
	if total == 0 {
		// Somebody using call as check.
		fmt.Printf("Call/Check")
		return player.TryPokerCheck(game, gc)
	}
	// Total bet by player:
	if ibet >= player.Chips {
		// Player is all-in
		ibet = player.Chips
		total = ibet + player.Bet
		if total <= curbet {
			fmt.Printf("Allin call\n")
			LogEvent(game, "Call", player.PlayerId, player.Chips, "ALL-IN")
			LogMessage(game, "%s goes all-in with %d", player.DisplayName, player.Chips)
			DoCall(game, tablename, gc.PlayerId, player.Chips)
			return SAVE|RUN
		}
		fmt.Printf("Allin bet\n")
		LogEvent(game, "Bet", player.PlayerId, player.Chips, "ALL-IN")
		LogMessage(game, "%s goes all-in with %d", player.DisplayName, player.Chips)
		// Else fall through to DoBet
	} else {
		// Player is not all-in.
		if total == table.CurBet {
			fmt.Printf("Call")
			// Call or Check
			DoCall(game, tablename, gc.PlayerId, ibet)
			LogEvent(game, "Call", player.PlayerId, ibet, "CALL")
			LogMessage(game, "%s calls with %d", player.DisplayName, ibet)
			return SAVE|RUN
		}
		// This is either a bet or a raise. Since player is not all-in,
		// this must be more than MinBet
		if (total - curbet) < table.MinBet {
			fmt.Printf("Bad bet")
			return ERR
		}
		if table.CurBet > 0 {
			LogEvent(game, "Bet", player.PlayerId, total - table.CurBet, "RAISE")
			LogMessage(game, "%s raises by %d", player.DisplayName, total - table.CurBet)
		} else {
			LogEvent(game, "Bet", player.PlayerId, total, "BET")
			LogMessage(game, "%s bets %d", player.DisplayName, total)
		}
	}

	DoBet(game, tablename, gc.PlayerId, int(ibet), false)
	return SAVE|RUN
}
