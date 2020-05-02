// Game DSLs for Poker-style games.
package rollpoker

import (
	"fmt"
	"time"
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

func (game *RoomData) Idle(_ string, _ int) bool {
	return true
}

func RunCommandInTransaction(game *RoomData, tablename string) time.Duration {
	table := game.Room.Tables[tablename]
	cmd := table.Dolist[0]

	method := reflect.ValueOf(game).MethodByName(cmd.Name)
	if !method.IsValid() {
		fmt.Printf("UNKNOWN COMMAND %s", cmd.Name)
		return -1
	}
	fmt.Printf("Running: %s\n", cmd.Name)

	args := []reflect.Value{ reflect.ValueOf(tablename), reflect.ValueOf(cmd.Arg) }
	rval := method.Call(args)
	table.Doing = cmd.Name
	if rval[0].Bool() {
		fmt.Printf("Popping: %s\n", cmd.Name)
		table.Dolist = table.Dolist[1:]
		if len(table.Dolist) > 0 {
			fmt.Printf("Next: %v\n", table.Dolist[0])
		}
		return cmd.Sleepfor
	}
	return -1
}

func RunCommandLoop(game *RoomData, tablename string) time.Duration {
	for {
		ret := RunCommandInTransaction(game, tablename)
		sanity := CheckGameSanity(game, ret >= 0)
		if sanity != "" {
			panic(sanity)
		}

		if ret != 0 {
			return ret
		}
	}
}

func RunCommandsTransaction(gamename string, tablename string) time.Duration {
	// RunCommand:
	//   1) Fetches a game, and pulls a table with Dolist instructions
	//   2) Runs the first command
	//   3) If it shouldn't run again (await user input / command returns false), returns -1
	//   4) If it should be run again (the command returns true):
	//   5) Pops first command off of game, then saves it.
	//   6) Returns a duration to sleep for.
	var ret time.Duration
	ret = -1
	count := 0
	err := FIRESTORE_CLIENT.RunTransaction(context.Background(),
					func(ctx context.Context, tx *firestore.Transaction) error {

		if count > 0 {
			return nil
		}
		count = count + 1

		game := FetchGame(gamename, tx)
		if (game == nil) {
			fmt.Println("FetchGame got a nil value")
			return nil
		}

		ret = RunCommandLoop(game, tablename)

		SaveGame(game, tx)
		return nil
	})
	if err != nil {
		fmt.Printf("ERROR in Transaction: %v\n", err)
	}
	return ret
}

func RunCommands(gamename string, tablename string, in_secs time.Duration) {
	if in_secs > 0 {
		time.Sleep(in_secs * time.Second)
	}
	result := RunCommandsTransaction(gamename, tablename)
	if result == 0 {
		fmt.Printf("Sleep of 0 should never be encountered in RunCommands")
	} else if result > 0 {
		go RunCommands(gamename, tablename, result)
	}
}

var GAME_COMMANDS map[string]GameDef

func init() {
	GAME_COMMANDS = make(map[string]GameDef)
	GAME_COMMANDS["texasholdem"] = GameDef{
		{"Idle", 0, 0}, // This gets tossed on new games.
		{"ResetHand", 0, 0},
		{"Shuffle", 0, 0},
		{"DealAllDown", 2, 0},
		{"ClearBets", 0, 0},
		{"HoldemBlinds", 0, 0},
		{"BetRound", 0, 0},
		{"CollectPot", 0, 0},
		{"Burn", 0, 0},
		{"TexFlop", 0, 1},
		{"ClearBets", 0, 0},
		{"BetRound", 0, 0},
		{"CollectPot", 0, 0},
		{"Burn", 0, 0},
		{"TexTurn", 0, 1},
		{"ClearBets", 0, 0},
		{"BetRound", 0, 0},
		{"CollectPot", 0, 0},
		{"Burn", 0, 0},
		{"TexRiver", 0, 1},
		{"ClearBets", 0, 0},
		{"BetRound", 0, 0},
		{"CollectPot", 0, 2},
		{"TexWin", 0, 8},
		{"BustOut", 0, 0},
		{"NewGame", 0, 0},
	}
	GAME_COMMANDS["_foldedwin"] = GameDef {
		{"Idle", 0, 0}, // This gets chopped off by return true
		{"CollectPot", 0, 0},
		{"FoldedWin", 0, 5},
		{"NewGame", 0, 0},
	}
	GAME_COMMANDS["_tourneywon"] = GameDef {
		{"Idle", 0, 0}, // This gets chopped off by return true
		{"GameWon", 0, 20},
		{"ClearGame", 0, 0},
	}
}
