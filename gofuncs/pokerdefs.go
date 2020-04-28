// Game DSLs for Poker-style games.
package rollpoker

import (
	"strings"
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

func (game *RoomData) Shuffle(tablename string, _ int) bool {
	DECK := []string{
		"sa", "s2", "s3", "s4", "s5", "s6", "s7", "s8", "s9", "st", "sj", "sq", "sk",
		"ha", "h2", "h3", "h4", "h5", "h6", "h7", "h8", "h9", "ht", "hj", "hq", "hk",
		"da", "d2", "d3", "d4", "d5", "d6", "d7", "d8", "d9", "dt", "dj", "dq", "dk",
		"ca", "c2", "c3", "c4", "c5", "c6", "c7", "c8", "c9", "ct", "cj", "cq", "ck",
	}

	LogEvent(game, "Shuffle")
	LogMessage(game, "Shuffling deck.")
	deckcopy := make([]string, len(DECK))
	copy(deckcopy, DECK)
	rand.Shuffle(len(deckcopy), func(i, j int) {
		deckcopy[i], deckcopy[j] = deckcopy[j], deckcopy[i]
	})
	tdata := FetchData(game, tablename)
	tdata.Cards = deckcopy
	SaveData(game, tablename, tdata)
	return true // We're done shuffling.
}

func (game *RoomData) Burn(tablename string, _ int) bool {
	tdata := FetchData(game, tablename)
	tdata.Cards = tdata.Cards[1:]
	SaveData(game, tablename, tdata)
	return true
}

func (game *RoomData) TexFlop(tablename string, _ int) bool {
	table := game.Room.Tables[tablename]
	tdata := FetchData(game, tablename)

	flop := tdata.Cards[:3]
	table.Cards["board"] = flop
	tdata.Cards = tdata.Cards[3:]
	LogEvent(game, "Board", strings.Join(flop, " "))
	LogMessage(game, "Flop: <<%s>>", strings.Join(flop, ">>,<<"))

	SaveData(game, tablename, tdata)
	return true
}

func (game *RoomData) TexTurn(tablename string, _ int) bool {
	table := game.Room.Tables[tablename]
	tdata := FetchData(game, tablename)

	table.Cards["board"] = append(table.Cards["board"], tdata.Cards[0])
	LogEvent(game, "Board", tdata.Cards[0])
	LogMessage(game, "Turn: <<%s>>", tdata.Cards[0])

	tdata.Cards = tdata.Cards[1:]

	SaveData(game, tablename, tdata)
	return true
}

func (game *RoomData) TexRiver(tablename string, _ int) bool {
	table := game.Room.Tables[tablename]
	tdata := FetchData(game, tablename)

	table.Cards["board"] = append(table.Cards["board"], tdata.Cards[0])
	LogEvent(game, "Board", tdata.Cards[0])
	LogMessage(game, "Turn: <<%s>>", tdata.Cards[0])

	tdata.Cards = tdata.Cards[1:]

	SaveData(game, tablename, tdata)
	return true
}

func (game *RoomData) BustOut(tablename string, _ int) bool {
	table := game.Room.Tables[tablename]
	busts := make([]*Player, 0)
	ranking := 0
	for _, table := range game.Room.Tables {
		ranking += len(table.Seats)
	}
	for seat, pid := range table.Seats {
		player := game.Room.Players[pid]
		if player.Chips == 0 {
			busts = append(busts, player)
			delete(table.Seats, seat)
		}
	}
	if len(busts) > 0 {
		sort.Slice(busts, func(i, j int) bool { return busts[i].TotalBet < busts[j].TotalBet })
		for _, player := range busts {
			LogEvent(game, "Bust", player.PlayerId, ranking)
			LogMessage(game, "%s busts out with rank: %d", player.DisplayName, ranking)
			player.Rank = ranking
			ranking--
			player.State = BUSTED
			player.DisplayState = "Busted out"
		}
	}
	return true
}

func (game *RoomData) NewGame(tablename string, _ int) bool {
	table := game.Room.Tables[tablename]
	table.Dolist = make(GameDef, len(GAME_COMMANDS["texasholdem"]))
	table.Dealer = GetNextPlayer(game, table, table.Dealer)
	copy(table.Dolist, GAME_COMMANDS["texasholdem"])
	LogEvent(game, "Newhand")
	LogMessage(game, "New Hand")
	return true
}

func (game *RoomData) DealAllDown(tablename string, count int) bool {
	// Just for kicks, though our users will never know, let's deal it as if
	// it's a real poker game.
	table := game.Room.Tables[tablename]
	order := GetNextPlayers(game, table, table.Dealer)

	tdata := FetchData(game, tablename)

	var idx = 0

	for i := 0; i < count; i++ {
		for _, seat := range order {
			playerid := table.Seats[seat]
			player := game.Room.Players[playerid]
			phand := FetchData(game, playerid)
			player.Hand = append(player.Hand, "bg")
			phand.Cards = append(phand.Cards, tdata.Cards[idx])
			SaveData(game, playerid, phand)
			idx += 1
		}
	}
	tdata.Cards = tdata.Cards[idx:]
	LogEvent(game, "Dealing", count, "DOWN")
	LogMessage(game, "Dealing %d cards to each player", count)

	SaveData(game, tablename, tdata)

	return true
}

type Pot struct {
	Players		[]*Player
	BetAmount	int
	Chips		int
	WinningScore	int
	Winners		[]*Player
}

type PlayerHand struct {
	Player	*Player
	Hand	string
	Cards	[]string
	Score	int
}

func MakePots(game *RoomData, table *TableState) []*Pot {
	stillins := make([]*Player, 0)
	seenpots := make(map[int]bool)
	table.Pot = 0
	for _, pid := range table.Seats {
		player := game.Room.Players[pid]
		if player.State != FOLDED {
			stillins = append(stillins, player)
			seenpots[player.TotalBet] = true
		}
		player.Bet = 0
	}
	fmt.Printf("# of seenpots: %d\n", len(seenpots))
	pots, idx := make([]*Pot, len(seenpots)), 0
	for amt, _ := range seenpots {
		pots[idx] = new(Pot)
		pots[idx].Players = make([]*Player, 0)
		pots[idx].BetAmount = amt
		idx++
	}

	fmt.Printf("# of pots: %d\n", len(pots))
	taken := 0
	sort.Slice(pots, func(i, j int) bool { return pots[i].BetAmount < pots[j].BetAmount })
	for idx = 0; idx < len(pots); idx++ {
		for _, pid := range table.Seats {
			player := game.Room.Players[pid]
			// We can get from (taken) to (max) chips
			max := pots[idx].BetAmount
			if (max > player.TotalBet) {
			  max = player.TotalBet
			}
			if max > taken {
				pots[idx].Chips += max - taken

				if player.State == CALLED || player.State == ALLIN || player.State == BET {
					pots[idx].Players = append(pots[idx].Players, player)
				}
			}
		}
		taken = pots[idx].BetAmount
	}
	fmt.Println("Pots:")
	fmt.Println(pots)
	return pots
}

func PayoutPots(game *RoomData, pots []*Pot, hands []*PlayerHand) {
	// Sort in reverse, so greatest first
	sort.Slice(hands, func(i, j int) bool { return hands[i].Hand > hands[j].Hand })
	// Map PlayerId to Hands
	idScore := make(map[string]*PlayerHand)
	for _, hand := range hands { idScore[hand.Player.PlayerId] = hand }
	fmt.Println("Payout # of idscores: %v\n", len(idScore))

	// Find winning hand for each pot
	for i, _ := range pots {
		pots[i].WinningScore = 0
		for _, player := range pots[i].Players {
			if idScore[player.PlayerId].Score > pots[i].WinningScore {
				pots[i].WinningScore = idScore[player.PlayerId].Score
			}
		}
	}
	fmt.Println("Pots w/ winning scores")
	fmt.Println(pots)

	// Find winning players for each pot
	for i, _ := range pots {
		pots[i].Winners = make([]*Player, 0)
		for _, player := range pots[i].Players {
			if idScore[player.PlayerId].Score == pots[i].WinningScore {
				ph := idScore[player.PlayerId]
				pots[i].Winners = append(pots[i].Winners, player)
				// We have a winner, show this player's cards.
				player.Hand = GetPlayerHand(game, player.PlayerId)
				LogEvent(game, "Win", player.PlayerId, pots[i].Chips, ph.Hand, strings.Join(ph.Cards," "))
				LogMessage(game, "%s wins the %d-chip pot with %s: <<%s>>",
						 player.DisplayName, pots[i].Chips,
						 ph.Hand, strings.Join(ph.Cards, ">> <<"))
				player.DisplayState = idScore[player.PlayerId].Hand
			}
		}
	}

	// Pay out each pot
	for i, _ := range pots {
		DivvyPot(game, pots[i])
	}
}

func DivvyPot(game *RoomData, pot *Pot) {
	// pot.Winners contains the players that won.
	// pot.Chips contains how many chips to give out.
	fmt.Println("Pot to pay out:")
	fmt.Println(pot)
	if len(pot.Winners) == 1 {
		// Simple! Yaaaay!
		pot.Winners[0].Chips += pot.Chips
		return
	}

	// Crap. We have ties
	chipCount := int(pot.Chips / len(pot.Winners))

	total := 0
	for _, winner := range pot.Winners {
		winner.Chips += chipCount
		total += chipCount
	}
	pot.Winners[0].Chips += pot.Chips - total
}

func (game *RoomData) TexWin(tablename string, _ int) bool {
	// This is called at end of a hand, with at least 2 players in.
	// 1) Order still-in players by hand rank.
	// 2) For each still-in player, in order from best to worst,
	//    2a) Take their TotalBet chips from the rest's TotalBets.
	//    2b) Any player w/ 0 chips left from TotalBets is busted out.
	table := game.Room.Tables[tablename]
	pots := MakePots(game, table)
	allhands := make([]*PlayerHand, len(table.Seats))
	idx := 0

	// Determine winners, and set their State to what they had.
	for _, pid := range table.Seats {
		player := game.Room.Players[pid]
		allhands[idx] = new(PlayerHand)
		allhands[idx].Player = player
		hand := GetPlayerHand(game, pid)
		if player.State != FOLDED {
			allhands[idx].Cards, allhands[idx].Hand, allhands[idx].Score = GetTexasRank(hand, table.Cards["board"])
		} else {
			allhands[idx].Hand = ""
			allhands[idx].Score = 0
		}
		idx++
	}

	PayoutPots(game, pots, allhands)
	return true
}

func (game *RoomData) FoldedWin(tablename string, _ int) bool {
	// There should only be one active player when this is called.
	// All others should be FOLDED
	table := game.Room.Tables[tablename]
	active := []*Player{}
	for _, playerid := range table.Seats {
		player := game.Room.Players[playerid]

		if player.State != FOLDED {
			active = append(active, player)
		}
	}
	if len(active) != 1 {
		fmt.Printf("ERROR: FoldedWin with %d active?", len(active))
		return false
	}
	player := active[0]
	LogEvent(game, "Win", player.PlayerId, table.Pot)
	LogMessage(game, "%s wins %d chips", player.DisplayName, table.Pot)
	player.Chips += table.Pot
	player.State = WON
	player.DisplayState = "Winner"
	table.Pot = 0
	return true
}

func GetPlayerHand(rdata *RoomData, playerid string) []string {
	phand := FetchData(rdata, playerid)
	return phand.Cards
}

func (game *RoomData) ResetHand(tablename string, _ int) bool {
	table := game.Room.Tables[tablename]
	table.Pot = 0
	for _, playerid := range table.Seats {
		game.Room.Players[playerid].Bet = 0
		game.Room.Players[playerid].TotalBet = 0
		game.Room.Players[playerid].Hand = make([]string, 0)
		emptyHand := new(DataItem)
		emptyHand.Cards = []string{}
		SaveData(game, playerid, emptyHand)
		game.Room.Players[playerid].State = WAITING
		game.Room.Players[playerid].DisplayState = ""
		table.Cards = make(map[string][]string)
	}
	return true
}

func (game *RoomData) TableFor(playerid string) string {
	for tn, table := range game.Room.Tables {
		for _, pid := range table.Seats {
			if playerid == pid {
				return tn
			}
		}
	}
	return ""
}

func (game *RoomData) TableForPlayer(player *Player) string {
	return game.TableFor(player.PlayerId)
}

func DoChoose(game *RoomData, tablename, playerid, state, dstate string) {
	player := game.Room.Players[playerid]
	player.State = state
	player.DisplayState = dstate
}

func DoFold(game *RoomData, tablename, playerid string) {
	player := game.Room.Players[playerid]
	player.Hand = make([]string, 0)
	DoChoose(game, tablename, playerid, FOLDED, "Folded")
}

func DoCall(game *RoomData, tablename, playerid string, amt int) {
	player := game.Room.Players[playerid]
	table := game.Room.Tables[tablename]
	if player.Chips <= amt {
		amt = player.Chips
	}
	if amt == 0 {
		DoChoose(game, tablename, playerid, CALLED, "Checked")
	} else {
		DoChoose(game, tablename, playerid, CALLED, "Called")
	}
	player.Chips -= amt
	player.Bet += amt
	player.TotalBet += amt
	if player.Bet > table.CurBet {
		fmt.Printf("How did player.Bet > table.CurBet in DoCall?")
	}
}

func DoBet(game *RoomData, tablename, playerid string, amt int, auto bool) {
	player := game.Room.Players[playerid]
	table := game.Room.Tables[tablename]
	// Reset all other CALLED and BET players' states to WAITING
	if !auto {
		for _, pid := range table.Seats {
			if (game.Room.Players[pid].State == CALLED || game.Room.Players[pid].State == BET) {
				if (game.Room.Players[pid].Chips == 0) {
					game.Room.Players[pid].State = ALLIN
				} else {
					game.Room.Players[pid].State = WAITING
					// No change to other players' display states.
				}
			}
		}
	}
	if player.Chips <= amt {
		amt = player.Chips
	}
	if !auto {
		// auto is true for blinds
		DoChoose(game, tablename, playerid, BET, "Bet")
	}
	player.Chips -= amt
	player.Bet += amt
	player.TotalBet += amt
	diff := amt - table.CurBet
	if diff > table.MinBet {
		table.MinBet = diff
	}
	if player.Bet > table.CurBet {
		table.CurBet = player.Bet
	}
	if auto && player.Chips <= 0 {
		player.State = ALLIN
	}
	if player.Chips == 0 {
		player.DisplayState = "All-In"
	}
}

func (game *RoomData) HoldemBlinds(tablename string, _ int) bool {
	// Called on game start: Make first two players to left of dealer
	// do their blind bets, set all players WAITING (if not ALLIN)
	table := game.Room.Tables[tablename]
	table.Pot = 0
	order := GetNextPlayers(game, table, table.Dealer)
	if len(order) < 2 { return false }
	names := []string{"(small blind)", "(big blind)", "(blind)"}
	for idx, seat := range order {
		playerid := table.Seats[seat]
		if idx < len(game.Room.CurrentBlinds) {
			// DoBet sets ALLIN if needed
			player := game.Room.Players[playerid]
			if idx < len(names) {
				player.DisplayState = names[idx]
			} else {
				player.DisplayState = names[len(names)-1]
			}
			DoBet(game, tablename, playerid, game.Room.CurrentBlinds[idx], true)
		} else {
			break
		}
	}
	table.CurBet = game.Room.CurrentBlinds[len(game.Room.CurrentBlinds)-1]
	table.MinBet = game.Room.CurrentBlinds[len(game.Room.CurrentBlinds)-1]
	return true
}

func (game *RoomData) ClearBets(tablename string, _ int) bool {
	// ClearBets does three things:
	// 1) Sets any active players WAITING
	// 2) Sets any active players with 0 chips ALLIN
	// 3) Sets table.MinBet and table.CurBet
	table := game.Room.Tables[tablename]
	for _, pid := range table.Seats {
		player := game.Room.Players[pid]
		if player.State != FOLDED {
			if player.Chips == 0 {
				player.State = ALLIN
				player.DisplayState = "All-In"
			} else {
				player.State = WAITING
				player.DisplayState = ""
			}
		}
	}
	table.MinBet = game.Room.CurrentBlinds[len(game.Room.CurrentBlinds) - 1]
	table.CurBet = 0
	return true
}

func (game *RoomData) BetRound(tablename string, _ int) bool {
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

	table := game.Room.Tables[tablename]

	dealorder := GetNextPlayers(game, table, table.Dealer)

	// Who has highest bet?
	betseat := table.Dealer
	betmin := 0

	for _, seat := range dealorder {
		pid := table.Seats[seat]
		player := game.Room.Players[pid]
		if player.Bet > betmin {
			betmin = player.Bet
			betseat = seat
		}
		if player.State == WAITING {
			waiting = append(waiting, pid)
		} else if player.State == BET || player.State == CALLED {
			called = append(called, pid)
			if player.State == BET {
				betseat = seat
			}
		} else if player.State == ALLIN {
			allins = append(allins, pid)
		} else if player.State == TURN {
			fmt.Println("How do we have a player with state TURN?")
		}
	}

	if (len(waiting) + len(called) + len(allins)) < 1 {
		fmt.Println("ERROR: How did waiting+called+allins get to be < 1?")
		return false
	} else if (len(waiting) + len(called) + len(allins)) == 1 {
		// All but one have folded. Short-circuit and that
		// player wins.
		table.Dolist = make(GameDef, len(GAME_COMMANDS["_foldedwin"]))
		copy(table.Dolist, GAME_COMMANDS["_foldedwin"])
		return true
	} else if (len(waiting) + len(called)) == 1 && len(allins) > 0 {
		// One or more players all-in. One non-allin player who has called or is waiting (start of new turn)
		// This is a little twisted - ALLIN is set only after a raise, or a turn has ended.
		allins = append(allins, called...)
		allins = append(allins, waiting...)
		washidden := false
		for _, playerid := range allins {
			player := game.Room.Players[playerid]
			for _, s := range player.Hand {
				if s == "bg" { washidden = true; break; }
			}
			if washidden {
				player.Hand = GetPlayerHand(game, playerid)
				LogMessage(game, "%s has: <<%s>>", player.DisplayName, strings.Join(player.Hand, ">>,<<"))
			}
		}
		return true
	} else if len(waiting) == 0 {
		// Nobody is waiting. Continue.
		return true
	}

	// Multiple waiters. Who plays?

	nextorder := GetNextPlayers(game, table, betseat)

	for _, seat := range nextorder {
		pid := table.Seats[seat]
		player := game.Room.Players[pid]
		if player.State == WAITING {
			player.State = TURN
			player.DisplayState = "Betting"
			return false
		}
	}
	fmt.Println("We should not get here...")
	return false
}

func (game *RoomData) Idle(_ string, _ int) bool {
	return true
}

func (game *RoomData) CollectPot(tablename string, _ int) bool {
	table := game.Room.Tables[tablename]
	amt := 0
	for _, pid := range table.Seats {
		player := game.Room.Players[pid]
		amt += player.Bet
		player.Bet = 0
	}
	table.Pot += amt
	return true
}

func GetNextPlayers(game *RoomData, table *TableState, seat string) []string {
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

func GetNextPlayer(game *RoomData, table *TableState, seat string) string {
	return GetNextPlayers(game, table, seat)[0]
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
		fmt.Printf("Next: %v\n", table.Dolist[0])
		fmt.Printf("Nextg: %v\n", game.Room.Tables[tablename].Dolist[0])
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
	err := FIRESTORE_CLIENT.RunTransaction(context.Background(),
					func(ctx context.Context, tx *firestore.Transaction) error {

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
		{"CollectPot", 0, 0},
		{"TexWin", 0, 8},
		{"BustOut", 0, 0},
		{"NewGame", 0, 0},
	}
	GAME_COMMANDS["_foldedwin"] = GameDef {
		{"Burn", 0, 0}, // This gets chopped off by return true
		{"CollectPot", 0, 0},
		{"FoldedWin", 0, 5},
		{"NewGame", 0, 0},
	}
}
