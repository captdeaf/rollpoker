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
	Members		map[string]string	// People who can view and sign up, and their display names.
	RoomPass	string		// Password to register as member.
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

type DataItem struct {
	// This is a catch-all for all game/gamename/data/<doc> documents.
	Cards []string
}

type DataRef struct {
	Data	*DataItem
	Doc	*firestore.DocumentSnapshot
	DocRef	*firestore.DocumentRef
	Changed	bool // If this should be updated on SaveGame or not.
}

type RoomData struct {
	// This structure doesn't exist in the DB. Instead, it's used
	// to pass around information on the server side. We fetch and
	// populate GameRoom from GameRoom DB entry.
	Name	string
	Room	GameRoom
	TX	*firestore.Transaction
	Logs	[]*LogItem
	Drefs	map[string]*DataRef
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

func FetchData(rdata *RoomData, name string) *DataItem {
	dref, has := rdata.Drefs[name]
	if !has {
		dref = new(DataRef)
		dref.Changed = false
		dref.DocRef = FIRESTORE_CLIENT.Doc("games/" + rdata.Name + "/data/" + name)
		doc, err := rdata.TX.Get(dref.DocRef)
		dref.Doc = doc
		dref.Data = new(DataItem)
		if err == nil && dref.Doc.Exists() {
			err = dref.Doc.DataTo(dref.Data)
		}
		dref.Changed = false
	}
	return dref.Data
}

func SaveData(rdata *RoomData, name string, data *DataItem) {
	dref, has := rdata.Drefs[name]
	if !has {
		dref = new(DataRef)
		rdata.Drefs[name] = dref
		dref.DocRef = FIRESTORE_CLIENT.Doc("games/" + rdata.Name + "/data/" + name)
	}
	dref.Data = data
	dref.Changed = true
}

func FetchGame(name string, tx *firestore.Transaction) *RoomData {
	var rdata RoomData
	pubRef := FIRESTORE_CLIENT.Doc("games/" + name)
	pub, err := tx.Get(pubRef)
	if err != nil {
		log.Printf("Can't get snapshot: %v\n", err)
		return nil
	}
	err = pub.DataTo(&rdata.Room)
	if err != nil {
		log.Printf("No pub.datato avail: %v\n", err)
		return nil
	}
	rdata.Name = name
	rdata.Drefs = make(map[string]*DataRef)
	rdata.TX = tx
	return &rdata
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
	for _, dref := range game.Drefs {
		if dref.Changed {
			game.TX.Set(dref.DocRef, dref.Data)
		}
	}
	fmt.Println("Saved", game.Name)
}

func CheckGameSanity(rdata *RoomData, hasCommandWaiting bool) string {
	if rdata.Room.RoomState != POKER {
		return ""
	}
	totalChips := 0
	for _, table := range rdata.Room.Tables {
		turncount := 0
		betcount := 0
		totalChips += table.Pot
		for _, pid := range table.Seats {
			player := rdata.Room.Players[pid]
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
	expectedChips := len(rdata.Room.Players) * rdata.Room.GameSettings.StartingChips

	if totalChips != expectedChips {
		return "Chip count mismatch!"
	}

	return ""
}

type CommandResponse struct {
	Errmsg	string	// If empty, command was a success.
	Run	bool	// If game should run commands (immediately)
	Save	bool	// If game should save
	Willrun	bool	// If command will deal with running commands. (For sanity check)
}

func CError(errmsg string) *CommandResponse {
	return CResponse(errmsg, false, false, false)
}

func CSave() *CommandResponse {
	return CResponse("", false, true, false)
}

func COK() *CommandResponse {
	return CResponse("", true, true, false)
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
	w.Header().Add("Content-Type", "text/plain")

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


	txerr := FIRESTORE_CLIENT.RunTransaction(context.Background(),
					func(ctx context.Context, tx *firestore.Transaction) error {

		rdata := FetchGame(gc.Name, tx)
		if rdata == nil {
			http.Error(w, "Unknown Game", http.StatusBadRequest)
			return nil
		}
		player = rdata.Room.Players[gc.PlayerId]

		fmt.Println("Got", gc.Command)

		// Call Command by name if it has one
		method := reflect.ValueOf(player).MethodByName("Try" + rdata.Room.RoomState + gc.Command)
		var cresp *CommandResponse

		if method.IsValid() {
			rval := method.Call([]reflect.Value{reflect.ValueOf(rdata), reflect.ValueOf(&gc)})
			cresp = rval[0].Interface().(*CommandResponse)
		} else {
			cresp = CError("Invalid command")
		}

		if cresp.Errmsg != "" {
			http.Error(w, cresp.Errmsg, http.StatusBadRequest)
			return nil
		}
		sanity := CheckGameSanity(rdata, cresp.Run || cresp.Willrun)
		if sanity != "" {
			http.Error(w, "Game failed sanity check", http.StatusBadRequest)
			panic(sanity)
		}
		var ret time.Duration
		var tbl string
		if cresp.Run {
			tbl = rdata.TableForPlayer(player)
			ret = RunCommandLoop(rdata, tbl)
		}
		if cresp.Save {
			SaveGame(rdata, tx)
		}
		if cresp.Run && ret >= 0 {
			go RunCommands(rdata.Name, tbl, ret)
		}
		fmt.Fprintf(w, "success\n")
		return nil
	})

	if txerr != nil {
		fmt.Printf("Error: %v\n", txerr)
	}
}
