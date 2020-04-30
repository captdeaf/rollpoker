// Player Commands for Poker
package rollpoker

import (
	"time"
	"fmt"
	"math/rand"
	"strings"
	"strconv"
)

// This is for Signup state, but it starts Poker state, so we include it here.
func (player *Player) TrySignupStartPoker(rdata *RoomData, gc *GameCommand) *CommandResponse {
	if (player == nil) { return CError("You are not registered to the tournament") }
	// Sanity checks:
	if len(rdata.Room.Players) < 2 { return CError("Not enough players") }
	// TODO: MTT support
	if len(rdata.Room.Players) > 10 { return CError("Too many players for now") }

	// So, we're trying to start a rdata.
	// 1) Do we have enough players? Do we need multiple tables?
	allPlayers := []string{}
	var settings GameSettings
	settings = *rdata.Room.OrigSettings
	rdata.Room.GameSettings = &settings

	for _, pl := range rdata.Room.Players {
		allPlayers = append(allPlayers, pl.PlayerId)
		pl.Chips = settings.StartingChips
		pl.Bet = 0
		pl.Rank = 0
		pl.State = ""
		pl.Hand = make([]string, 0)
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

	// 3) Populate TableState from GameSettings.
	rdata.Room.Tables = make(map[string]*TableState)
	table := TableState{}
	table.Seats = map[string]string{};
	for seatnum, pid := range allPlayers {
		table.Seats[allSeats[seatnum]] = pid
	}
	// First player is dealer, because player position is random, anyway.
	table.Dealer = ""
	rdata.Room.Tables["table0"] = &table
	table.Dolist = make(GameDef, len(GAME_COMMANDS["texasholdem"]))
	copy(table.Dolist, GAME_COMMANDS["texasholdem"])

	blindstr := rdata.Room.GameSettings.BlindStructure[0]
	if len(rdata.Room.GameSettings.BlindStructure) > 0 {
		rdata.Room.GameSettings.BlindStructure = rdata.Room.GameSettings.BlindStructure[1:]
		if len(rdata.Room.GameSettings.BlindTimes) > 0 {
			blindMinutes := rdata.Room.GameSettings.BlindTimes[0]
			if len(rdata.Room.GameSettings.BlindTimes) > 1 {
				rdata.Room.GameSettings.BlindTimes = rdata.Room.GameSettings.BlindTimes[1:]
			}
			rdata.Room.BlindTime = time.Now().Unix() + int64(60 * blindMinutes) // TODO: Now + blindTime.
		}
	}

	rdata.Room.PausedAt = 0
	rdata.Room.RoomState = POKER

	blindsplit := strings.Fields(blindstr)
	rdata.Room.CurrentBlinds = make([]int,len(blindsplit))
	for i, val := range blindsplit {
		ival, _ := strconv.ParseInt(val, 10, 32)
		rdata.Room.CurrentBlinds[i] = int(ival)
	}

	// We run the start for All tables by hand
	i := time.Duration(1)
	for tname, _ := range rdata.Room.Tables {
		go RunCommands(rdata.Name, tname, i)
		i += 1
	}
	fmt.Println("DName:", player.DisplayName)
	LogMessage(rdata, "%s starts the game", player.DisplayName)
	// We handle RunCommands, so we return special CommandResponse instead of COK
	return CResponse("", false, true, true)
}

func IsTurn(player *Player) bool {
	if player == nil { return false }
	if player.State != TURN { return false; }
	return true
}

func (player *Player) TryPokerKick(rdata *RoomData, gc *GameCommand) *CommandResponse {
	return COK()
}

func (player *Player) TryPokerCheck(rdata *RoomData, gc *GameCommand) *CommandResponse {
	if !IsTurn(player) { return CError("Not your turn") }
	tablename := rdata.TableForPlayer(player)
	table := rdata.Room.Tables[tablename]

	remaining := table.CurBet - player.Bet
	if remaining > 0 {
		return CError("You can't check.")
	}

	DoCall(rdata, tablename, gc.PlayerId, 0)
	LogEvent(rdata, "Call", player.PlayerId, remaining, "CHECK")
	LogMessage(rdata, "%s checks", player.DisplayName)
	return COK()
}

func (player *Player) TryPokerFold(rdata *RoomData, gc *GameCommand) *CommandResponse {
	if !IsTurn(player) { return CError("Not your turn") }
	tablename := rdata.TableForPlayer(player)

	DoFold(rdata, tablename, gc.PlayerId)
	LogEvent(rdata, "Fold", player.PlayerId, "FOLD")
	LogMessage(rdata, "%s folds", player.DisplayName)
	return COK()
}

func (player *Player) TryPokerCall(rdata *RoomData, gc *GameCommand) *CommandResponse {
	if !IsTurn(player) { return CError("Not your turn") }
	tablename := rdata.TableForPlayer(player)
	table := rdata.Room.Tables[tablename]

	remaining := table.CurBet - player.Bet
	if remaining > player.Chips {
		remaining = player.Chips
	}

	// Maximum is their amount of chips. (if chips < min, min = chips)
	DoCall(rdata, tablename, gc.PlayerId, remaining)
	if remaining == 0 {
		LogEvent(rdata, "Call", player.PlayerId, remaining, "CHECK")
		LogMessage(rdata, "%s checks", player.DisplayName)
	} else {
		LogEvent(rdata, "Call", player.PlayerId, remaining, "CALL")
		LogMessage(rdata, "%s calls for %d", player.DisplayName, remaining)
	}
	return COK()
}

func (player *Player) TryPokerBet(rdata *RoomData, gc *GameCommand) *CommandResponse {
	// This can be a Check, a Call, a Bet, or a Raise, depending on amount.
	if !IsTurn(player) { return CError("Not your turn") }
	tablename := rdata.TableForPlayer(player)
	table := rdata.Room.Tables[tablename]
	i64bet, ierr := strconv.ParseInt(gc.Args["amount"], 10, 32)
	ibet := int(i64bet)

	if ierr != nil || ibet < 0 { return CError("What do you want to bet?") }

	// Table current
	curbet := table.CurBet
	total := ibet + player.Bet
	if total == 0 {
		// Somebody using call as check.
		fmt.Printf("Call/Check")
		return player.TryPokerCheck(rdata, gc)
	}
	// Total bet by player:
	if ibet >= player.Chips {
		// Player is all-in
		ibet = player.Chips
		total = ibet + player.Bet
		if total <= curbet {
			fmt.Printf("Allin call\n")
			LogEvent(rdata, "Call", player.PlayerId, player.Chips, "ALL-IN")
			LogMessage(rdata, "%s goes all-in with %d", player.DisplayName, player.Chips)
			DoCall(rdata, tablename, gc.PlayerId, player.Chips)
			return COK()
		}
		fmt.Printf("Allin bet\n")
		LogEvent(rdata, "Bet", player.PlayerId, player.Chips, "ALL-IN")
		LogMessage(rdata, "%s goes all-in with %d", player.DisplayName, player.Chips)
		// Else fall through to DoBet
	} else {
		// Player is not all-in.
		if total == table.CurBet {
			fmt.Printf("Call")
			// Call or Check
			DoCall(rdata, tablename, gc.PlayerId, ibet)
			LogEvent(rdata, "Call", player.PlayerId, ibet, "CALL")
			LogMessage(rdata, "%s calls with %d", player.DisplayName, ibet)
			return COK()
		}
		// This is either a bet or a raise. Since player is not all-in,
		// this must be more than MinBet
		if (total - curbet) < table.MinBet {
			fmt.Printf("Bad bet")
			return CError(fmt.Sprintf("Need a minimum bet or raise of %d", table.MinBet))
		}
		if table.CurBet > 0 {
			LogEvent(rdata, "Bet", player.PlayerId, total - table.CurBet, "RAISE")
			LogMessage(rdata, "%s raises by %d", player.DisplayName, total - table.CurBet)
		} else {
			LogEvent(rdata, "Bet", player.PlayerId, total, "BET")
			LogMessage(rdata, "%s bets %d", player.DisplayName, total)
		}
	}

	DoBet(rdata, tablename, gc.PlayerId, int(ibet), false)
	return COK()
}
