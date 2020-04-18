package rollpoker

import (
	"fmt"
	"sort"
	"time"
	"math/rand"
	"context"
	"cloud.google.com/go/firestore"
	"reflect"
)

type GameCmd struct {
	Name string	// Name of function and call
	Arg int		// Argument for this event (e.g: "2" to deal 2 to each player)
	Sleepfor time.Duration	// How long to sleep after this event
}

type GameDef []GameCmd

func (game *Game) Shuffle(tablename string, _ int) bool {
	DECK := []string{
		"sa", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "st", "sj", "sq", "sk",
		"ha", "h2", "h3", "h4", "h5", "h6", "h7", "h8", "h9", "ht", "hj", "hq", "hk",
		"da", "d2", "d3", "d4", "d5", "d6", "d7", "d8", "d9", "dt", "dj", "dq", "dk",
		"ca", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "ct", "cj", "cq", "ck",
	}

	deckcopy := make([]string, len(DECK))
	copy(deckcopy, DECK)
	rand.Shuffle(len(deckcopy), func(i, j int) {
		deckcopy[i], deckcopy[j] = deckcopy[j], deckcopy[i]
	})
	game.Private.TableDecks[tablename] = deckcopy
	return true // We're done shuffling.
}

func (game *Game) DealAllDown(tablename string, count int) bool {
	// Just for kicks, though our users will never know, let's deal it as if
	// it's a real poker game.
	table := game.Public.Tables[tablename]
	order := GetNextPlayers(game, table, table.Dealer)
	deck := game.Private.TableDecks[tablename]

	var idx = 0

	for i := 0; i < count; i++ {
		for _, seat := range order {
			player := game.Public.Players[table.Seats[seat]]
			player.Hand = append(player.Hand, deck[idx])
			idx += 1
		}
	}
	game.Private.TableDecks[tablename] = deck[idx:]

	return true
}

func (game *Game) ResetHand(tablename string, _ int) bool {
	table := game.Public.Tables[tablename]
	table.Pot = 0
	for _, playerid := range table.Seats {
		game.Public.Players[playerid].Bet = 0
		game.Public.Players[playerid].Hand = []string{}
		game.Public.Players[playerid].State = WAITING
		table.Cards = make(map[string][]string)
	}
	return true
}

func DoBet(game *Game, tablename, playerid string, amt int, auto bool) {
	player := game.Public.Players[playerid]
	table := game.Public.Tables[tablename]
	// Reset all other CALLED and BET players' states to WAITING
	if !auto {
		for _, pid := range table.Seats {
			if (game.Public.Players[pid].State == CALLED ||
			    game.Public.Players[pid].State == BET) {
				game.Public.Players[pid].State = WAITING
			}
		}
	}
	if player.Chips <= amt {
		amt = player.Chips
		game.Public.Players[playerid].State = ALLIN
	} else if !auto {
		game.Public.Players[playerid].State = BET
	}
	player.Chips -= amt
	player.Bet = amt
}

func (game *Game) HoldemBlinds(tablename string, _ int) bool {
	// Called on game start: Make first two players to left of dealer
	// do their blind bets, set all players WAITING (if not ALLIN)
	table := game.Public.Tables[tablename]
	table.Pot = 0
	order := GetNextPlayers(game, table, table.Dealer)
	if len(order) < 2 { return false }
	firstid := ""
	for idx, seat := range order {
		playerid := table.Seats[seat]
		game.Public.Players[playerid].State = WAITING
		if idx < len(game.Public.CurrentBlinds) {
			// DoBet sets ALLIN if needed
			DoBet(game, tablename, playerid, game.Public.CurrentBlinds[idx], true)
		} else {
			if firstid == "" {
				firstid = playerid
			}
			game.Public.Players[playerid].Bet = 0
		}
		// All players should be WAITING (or ALLIN), except firstid
	}
	if firstid == "" {
		firstid = order[0]
	}
	game.Public.Players[firstid].State = TURN
	return true
}

func (game *Game) BetRound(tablename string, isfirst int) bool {
	// BetRound() is called at the start of a betting round, and when any
	// bet, call, fold, etc is made. Simple check: If any WAITING,
	// or only one WAITING and the rest ALLIN or FOLDED, we return true
	// so the game can continue to the next step.
	// Otherwise advance to next player and return false.
	//
	// Special case: Only one active // canbet, and no allins: They win, we
	// shortcut Table.commands.
	waiting := []string{} // Players with WAITING status
	allins := []string{}  // Players with ALLIN status. Can't bet, but still in game
	called := []string{}  // players with BET or CALLED status

	table := game.Public.Tables[tablename]
	for _, pid := range table.Seats {
		player := game.Public.Players[pid]
		if player.State == WAITING || player.State == TURN {
			waiting = append(waiting, pid)
		} else if (player.State == BET || player.State == CALLED) {
			called = append(called, pid)
		} else if (player.State == ALLIN) {
			allins = append(allins, pid)
		}
	}
	if (len(waiting) + len(called) + len(allins)) == 1 {
		// All but one have folded. Short-circuit and that
		// player wins.
		// TODO: Replace command set with ClosedWin and
		// continue game
		return true
	} else if len(waiting) == 0 && len(called) == 1 && len(allins) > 0 {
		// One or more players all-in. Other player called.
		// No more need for bets.
		return true
	} else if len(waiting) == 1 && len(called) == 0 && len(allins) > 0 {
		// One waiter, and one or more allins.
		// Two instances:
		// 1) Player A all-ins, player B needs to call
		// 2) B's called A. Flop is dealt, now WAITING again.
		// If (1): return false, if (2): return true
		for _, pid := range allins {
			player := game.Public.Players[pid]
			if player.Bet > 0 {
				return false
			}
		}
		return true
	}
	return len(waiting) == 0
}

func (game *Game) CollectPot(tablename string, isfirst int) bool {
	table := game.Public.Tables[tablename]
	amt := 0
	for _, pid := range table.Seats {
		player := game.Public.Players[pid]
		amt += player.Bet
		player.Bet = 0
	}
	table.Pot += amt
	return true
}

func GetNextPlayers(game *Game, table *TableState, seat string) []string {
	// We have a table to deal out to. Table has Seats, in string order.
	// We pull them out, order by seat#, then rotate Dealer, and rotate
	// array around Dealer
	allseats := make([]string, 0, len(table.Seats))
	for seat, _ := range table.Seats {
		allseats = append(allseats, seat)
	}
	sort.Strings(allseats)

	nextidx := -1
	for idx, val := range(allseats) {
		if val > seat && nextidx < 0 {
			nextidx = idx
			break
		}
	}
	var orderedSeats []string
	if nextidx <= 0 {
		orderedSeats = allseats
	} else {
		orderedSeats = append(allseats[nextidx:len(allseats)], allseats[0:nextidx]...)
	}
	return orderedSeats
}

func GetNextPlayer(game *Game, table *TableState, seat string) string {
	return GetNextPlayers(game, table, seat)[0]
}

func RunCommandTransaction(gamename string, tablename string) time.Duration {
	// RunCommand:
	//   1) Fetches a game, and pulls a table with Dolist instructions
	//   2) Runs the first command
	//   3) If it shouldn't run again (await user input / command returns false), returns -1
	//   4) If it should be run again (the command returns true):
	//   5) Pops first command off of game, then saves it.
	//   6) Returns a duration to sleep for.
	var ret time.Duration
	ret = -1
	FIRESTORE_CLIENT.RunTransaction(context.Background(),
					func(ctx context.Context, tx *firestore.Transaction) error {

		game := FetchGame(gamename, tx)

		table := game.Public.Tables[tablename]
		cmd := table.Dolist[0]

		method := reflect.ValueOf(game).MethodByName(cmd.Name)
		if !method.IsValid() {
			fmt.Printf("UNKNOWN COMMAND %s", cmd.Name)
			return nil
		}

		args := []reflect.Value{ reflect.ValueOf(tablename), reflect.ValueOf(cmd.Arg) }
		rval := method.Call(args)
		docontinue := rval[0].Bool()
		table.Doing = cmd.Name
		if docontinue {
			table.Dolist = table.Dolist[1:]
			ret = cmd.Sleepfor
		}
		SaveGame(game, tx)
		return nil
	});
	return ret
}

func RunCommands(gamename string, tablename string, in_secs time.Duration) {
	if in_secs > 0 {
		time.Sleep(in_secs * time.Second)
	}
	result := RunCommandTransaction(gamename, tablename)
	if result >= 0 {
		go RunCommands(gamename, tablename, result)
	}
}

var GAME_COMMANDS map[string]GameDef

func init() {
	GAME_COMMANDS = map[string]GameDef{
		"texasholdem": {
			{"ResetHand", 0, 0},
			{"Shuffle", 0, 2},
			{"DealAllDown", 2, 2},
			{"HoldemBlinds", 0, 2},
			{"BetRound", 1, 2},
			{"CollectPot", 0, 2},
			{"Burn", 0, 2},
			{"TexFlop", 0, 2},
			{"BetRound", 0, 2},
			{"CollectPot", 0, 2},
			{"Burn", 0, 2},
			{"TexTurn", 0, 2},
			{"BetRound", 0, 2},
			{"CollectPot", 0, 2},
			{"Burn", 0, 2},
			{"TexRiver", 0, 2},
			{"BetRound", 0, 2},
			{"CollectPot", 0, 2},
			{"Texwin", 0, 5},
		},
	}
}
