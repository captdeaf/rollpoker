package rollpoker

import (
	"reflect"
	"os"
	"fmt"
	"time"
	"math/rand"
	"context"
	"log"
	"strconv"
	"strings"
	"net/http"
	"encoding/json"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
)

var BASE_URI string = "https://rollpoker.web.app"
var SEND_EMAIL bool = true
var FAKE_COMMANDS bool = false

type GameSettings struct {
	GameType	string	// Cash, SitNGo
	BetLimit	string	// NoLimit, PotLimit
	StartingChips	int	// 1500
	BlindStructure	[]string	// 25 50,25 75,50 100,75 150,...
	BlindTimes	[]int	// 40 40 40 20, for first 3 to be 40 minutes, then 20 mins after that.
}

type Player struct {
	PlayerId	string		// MadPinkWhale
	DisplayName	string		// "Bob D"
	DisplayState	string		// "Active" "Zzzzz" "Not connected"
	Rank		int		// Assigned on Bust out or Game won
	Chips		int		// 1500
	Bet		int		// Current amount bet // inside the circle
	TotalBet	int		// Running total
	State		string		// "Waiting" "Folded" etc
	Hand		[]string	// "hasa" - decrypted, or "!<string>" encrypted
}

type TableState struct {
	Seats		map[string]string	// seat0...seat9 to PlayerId
	Pot		int		// 200 ... total in pot, but not in bets
	Dealer		string		// seat0...seat9
	Dolist		GameDef		// GAME_COMMANDS["texasholdem"], etc
	Cards		map[string][]string // "flop": ["ha", "hk", "hq"], "turn": ...
	Doing		string		// Command name, in case we can reflect it client-side
	MinBet		int		// Big blind, or high bet so far.
	CurBet		int		// Sum of all bets and raises so far
}

type GameRoom struct {
	RoomState	string	// "SIGNUP" "POKER"
	GameSettings	*GameSettings
	OrigSettings	*GameSettings
	Tables		map[string]*TableState	// table0: ...
	Players		map[string]*Player	// UID: Player
	CurrentBlinds	[]int
	BlindTime	int64
	PausedAt	int64		// Nonzero if paused
	Password	string		// Password to register as member.
}

type LogItem struct {
	PlayerId	string
	Message		string
	EventName	string
	Args		[]interface{}
}

type LogItems struct {
	Timestamp	int64
	Logs		*[]*LogItem
}

type RoomData struct {
	// This structure doesn't exist in the DB. Instead, it's used
	// to pass around information on the server side. We fetch and
	// populate GameRoom from GameRoom DB entry.
	Name	string
	Room	GameRoom
	TX	*firestore.Transaction
	Logs	[]*LogItem
}

const (
	BUSTED = "BUSTED"
	FOLDED = "FOLDED"	// Out of this hand
	TURN = "TURN"		// It's their turn to do something
	WAITING = "WAITING"	// Hasn't had their turn yet.
	BET = "BET"		// Bet or raised
	CALLED = "CALLED"	// All players but one must be CALLED (or ALLIN)
	ALLIN = "ALLIN"		// Has no more chips left.
	WON = "WON"		// Won the hand
)

const (
	SAVE = 0x01
	RUN = 0x02
	ERR = 0
)

func AddEvent(game *RoomData, playerid string, event string, args interface{}, logmsg string) {
}

// values for Public.RoomState
const (
	SIGNUP	= "Signup"
	POKER	= "Poker"
)

type GameCommand struct {
	Name		string
	PlayerId	string
	Command		string
	Args		map[string] string
	ErrorMessage	string
}

var FIRESTORE_CLIENT *firestore.Client

func init() {
	// Seed the random number generator.
	rand.Seed(time.Now().UnixNano())
	ctx := context.Background()
	var err error
	_, err = os.Stat("firebase-key.json")
	if err == nil {
		opt := option.WithCredentialsFile("firebase-key.json")
		FIRESTORE_CLIENT, err = firestore.NewClient(ctx, "rollpoker", opt)
	} else {
		FIRESTORE_CLIENT, err = firestore.NewClient(ctx, "rollpoker")
	}
	if err != nil {
		log.Printf("Can't get client: %v\n", err)
		return
	}
	fmt.Println("Rollpoker started")
}

type GameResponse struct {
	Name string
}

func MakeTable(w http.ResponseWriter, r *http.Request) {
	var args map[string]string

	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("args: %v\n", args)

	var newgame RoomData
	var settings GameSettings

	settings.GameType = args["GameType"]
	settings.BetLimit = args["BetLimit"]
	i64, _ := strconv.ParseInt(args["StartingChips"], 10, 32)
	settings.StartingChips = int(i64)
	allblinds := strings.Split(args["BlindStructure"], ",")
	settings.BlindStructure	= make([]string, len(allblinds))	// 25 50,25 75,50 100,75 150,...

	for i, val := range allblinds {
		settings.BlindStructure[i] = val
	}

	alltimes := strings.Fields(args["BlindTimes"])
	settings.BlindTimes = make([]int, len(alltimes))
	for i, val := range alltimes {
		t64, _ := strconv.ParseInt(val, 10, 32)
		settings.BlindTimes[i] = int(t64)
	}

	newgame.Name = GenerateName()

	newgame.Room.Tables = make(map[string]*TableState)
	newgame.Room.RoomState = SIGNUP
	newgame.Room.OrigSettings = &settings

	FIRESTORE_CLIENT.RunTransaction(context.Background(),
					func(ctx context.Context, tx *firestore.Transaction) error {
		SaveGame(&newgame, tx)
		return nil
	})

	var newgr GameResponse
	newgr.Name = newgame.Name

	bytes, err := json.Marshal(newgr)
	w.Write(bytes)
}

func FetchData(game *RoomData, name string, ptr interface{}) bool {
	docRef := FIRESTORE_CLIENT.Doc("games/" + game.Name + "/data/" + name)
	doc, err := game.TX.Get(docRef)
	if err != nil {
		log.Printf("Can't get snapshot: %v\n", err)
		return false
	}
	err = doc.DataTo(ptr)
	if err != nil {
		log.Printf("No doc.datato avail: %v\n", err)
		return false
	}
	return true
}

func SaveData(game *RoomData, name string, ptr interface{}) {
	docRef := FIRESTORE_CLIENT.Doc("games/" + game.Name + "/data/" + name)
	err := game.TX.Set(docRef, ptr)
	if err != nil {
		log.Printf("Unable to save game data: %v", err)
	}
}

func FetchGame(name string, tx *firestore.Transaction) *RoomData {
	var game RoomData
	pubRef := FIRESTORE_CLIENT.Doc("games/" + name)
	pub, err := tx.Get(pubRef)
	if err != nil {
		log.Printf("Can't get snapshot: %v\n", err)
		return nil
	}
	err = pub.DataTo(&game.Room)
	if err != nil {
		log.Printf("No pub.datato avail: %v\n", err)
		return nil
	}
	game.Name = name
	game.TX = tx
	return &game
}

func LogEvent(game *RoomData, name string, fargs ...interface{}) {
	litem := new(LogItem)
	litem.Message = ""
	litem.EventName = name
	litem.Args = fargs
	fmt.Println(litem.Message)
	game.Logs = append(game.Logs, litem)
}

func LogMessage(game *RoomData, msg string, fargs ...interface{}) {
	litem := new(LogItem)
	litem.Message = fmt.Sprintf(msg, fargs...)
	fmt.Println(litem.Message)
	game.Logs = append(game.Logs, litem)
}

func SaveGame(game *RoomData, tx *firestore.Transaction) {
	pubref := FIRESTORE_CLIENT.Doc("games/" + game.Name)
	err := tx.Set(pubref, game.Room)
	if err != nil {
		fmt.Printf("Error saving games: %v\n", err)
		return
	}

	if len(game.Logs) > 0 {
		litems := new(LogItems)
		litems.Timestamp = time.Now().UnixNano()
		litems.Logs = &game.Logs

		lname := fmt.Sprintf("games/%s/log/%d", game.Name, litems.Timestamp)
		docref := FIRESTORE_CLIENT.Doc(lname)
		err = game.TX.Set(docref, litems)
		if err != nil {
			fmt.Printf("Error saving Log Items: %v\n", err)
			return
		}
	}
	fmt.Println("Saved", game.Name)
}

func CheckGameSanity(game *RoomData, hasCommandWaiting bool) string {
	if game.Room.RoomState != POKER {
		return ""
	}
	totalChips := 0
	for _, table := range game.Room.Tables {
		turncount := 0
		betcount := 0
		totalChips += table.Pot
		for _, pid := range table.Seats {
			player := game.Room.Players[pid]
			if player.State == TURN {
				turncount += 1
			}
			if player.State == BET {
				betcount += 1
			}
			totalChips += player.Bet
			totalChips += player.Chips
		}
		if turncount > 1 {
			return "Multiple players at one table have TURN"
		}
		if betcount > 1 {
			return "Multiple players at one table have BET"
		}
		if turncount != 1 && !hasCommandWaiting {
			return "No TURN or Command going??"
		}
	}
	expectedChips := len(game.Room.Players) * game.Room.GameSettings.StartingChips

	if totalChips != expectedChips {
		return "Chip count mismatch!"
	}

	return ""
}

func RegisterAccount(game *RoomData, gc *GameCommand) bool {
	return true
}

type CommandResponse struct {
	Errmsg	string	// If empty, command was a success.
	Run	bool	// If game should run commands (immediately)
	Save	bool	// If game should save
	Willrun	bool	// If command will deal with running commands. (For sanity check)
}

func CError(errmsg string) *CommandResponse {
	ret := new(CommandResponse)
	ret.Errmsg = errmsg
	ret.Run = false
	ret.Save = false
	ret.Willrun = false
	return ret
}

func COK() *CommandResponse {
	ret := new(CommandResponse)
	ret.Errmsg = ""
	ret.Run = true
	ret.Save = true
	ret.Willrun = true
	return ret
}

func CResponse(errmsg string, run, save, willrun bool) *CommandResponse {
	ret := new(CommandResponse)
	ret.Errmsg = errmsg
	ret.Run = run
	ret.Save = save
	ret.Willrun = willrun
	return ret
}

func Poker(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	// Only authorized players can run GameCommand commands.
	var uid = GetUserIDFromHeader(r)
	if uid == "" {
		return
	}

	var gc GameCommand
	var player *Player = nil
	err := json.NewDecoder(r.Body).Decode(&gc)

	if err != nil {
		return
	}
	gc.PlayerId = uid
	gc.ErrorMessage = ""

	dorun := false
	dosave := false

	txerr := FIRESTORE_CLIENT.RunTransaction(context.Background(),
					func(ctx context.Context, tx *firestore.Transaction) error {

		game := FetchGame(gc.Name, tx)
		if game == nil {
			http.Error(w, "Unknown Game", http.StatusBadRequest)
			return nil
		}

		fmt.Println("Got", gc.Command)

		if gc.Command == "invite" {
			if !RegisterAccount(game, &gc) {
				http.Error(w, "Unable to register", http.StatusBadRequest)
				return nil
			}
			dosave = true
			dorun = false
		} else {
			if player == nil {
				// If player is nil, the only thing the player
				// can do is "register" / invite.
				http.Error(w, "You are not a player", http.StatusBadRequest)
				return nil
			} else {
				// Call Command by name if it has one
				method := reflect.ValueOf(player).MethodByName("Try" + game.Room.RoomState + gc.Command)

				if method.IsValid() {
					rval := method.Call([]reflect.Value{reflect.ValueOf(game), reflect.ValueOf(&gc)})
					iret := rval[0].Int()
					dorun = iret & RUN == RUN
					dosave = iret & SAVE == SAVE
				} else {
					return nil
				}
			}
		}
		if gc.ErrorMessage != "" {
			http.Error(w, gc.ErrorMessage, http.StatusBadRequest)
		} else if dosave == false {
			http.Error(w, "You can't do that", http.StatusBadRequest)
		} else {
			fmt.Fprintf(w, "success")
		}
		sanity := CheckGameSanity(game, dorun || gc.Command == "StartGame")
		if sanity == "" {
			var ret time.Duration
			var tbl string
			if dorun {
				tbl = game.TableForPlayer(player)
				ret = RunCommandLoop(game, tbl)
			}
			if dosave {
				SaveGame(game, tx)
			}
			if dorun && ret >= 0 {
				go RunCommands(game.Name, tbl, ret)
			}
		} else {
			panic(sanity)
		}
		return nil
	})

	if txerr != nil {
		fmt.Printf("Error: %v\n", txerr)
	}
}
