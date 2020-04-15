package rollpoker

import (
	"reflect"
	"os"
	"fmt"
	"context"
	"log"
	"strconv"
	"strings"
	"net/http"
	"encoding/json"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type GameSettings struct {
	GameType	string	// Cash, SitNGo
	BetLimit	string	// NoLimit, PotLimit
	StartingChips	int	// 1500
	ChipValues	string	// White Red Blue Green Black Yellow: "25 100 500 1000..."
	BlindStructure	[]string	// 25 50,25 75,50 100,75 150,...
	BlindTimes	[]int	// 40 40 40 20, for first 3 to be 40 minutes, then 20 mins after that.
}

type PrivateGameInfo struct {
	PlayerKeys	map[string]string	// KillerOrangeHouse:FooBarBaz
	TableDecks	map[string][]string	// "table0": ["ha", "s3", ...], ...
	AdminPassword	string			// "hunter2"
	OrigState	GameSettings		// Original game state for restart/new game/etc
}

type Player struct {
	PlayerId	string		// MadPinkWhale
	DisplayName	string		// "Bob D"
	DisplayState	string		// "Active" "Zzzzz" "Not connected"
	Rank		int		// Assigned on Bust out or Game won
	Chips		int		// 1500
	Bet		int		// Current amount bet // inside the circle
	State		string		// "Waiting" "Folded" etc
}

type TableState struct {
	Seats		map[string]string	// seat0...seat9 to PlayerId
	Pot		int		// 200 ... total in pot, but not in bets
	Dealer		string		// seat0...seat9
}

type PublicGameInfo struct {
	State		string	// "NOGAME", "CASH", "SITNGO", etc.
	GameSettings	GameSettings
	Tables		map[string]TableState
	Players		map[string]Player
}

type GameEvent struct {
	EventId		int	// 0..9
	PlayerId	string	// MadOrangeCow
	Event		string	// Name for description e.g: "Fold"
	Args		string	// Parameters. "playerid"
	Log		string	// "Chris folded". If nil, no log.
}

type Game struct {
	Name	string
	Private	PrivateGameInfo
	Public	PublicGameInfo
}

const (
	BUSTED = "BUSTED"	// Never changes unless restart or cash buyin.
	FOLDED = "FOLDED"	// Out of this hand

	TURN = "TURN"		// It's their turn to do something
	WAITING = "WAITING"	// Hasn't had their turn yet.
	BET = "BET"		// Bet or raised
	CALLED = "CALLED"	// All players but one must be CALLED (or ALLIN)
	ALLIN = "ALLIN"		// Has no more chips left.
)

func AddEvent(game *Game, playerid string, event string, args interface{}, logmsg string) {
}

// values for Public.State
const (
	NOGAME = "NOGAME"	// Default
	CASHGAME = "CASHGAME"	// Cash in, cash out. Buy in or play
	SITNGO = "SITNGO"	// Active Sit-N-Go going
)

type GameCommand struct {
	Name		string
	PlayerId	string
	PlayerKey	string
	Command		string
	Args		map[string] string
}

var client *firestore.Client

func init() {
	ctx := context.Background()
	var err error
	_, err = os.Stat("firebase-key.json")
	if err == nil {
		opt := option.WithCredentialsFile("firebase-key.json")
		client, err = firestore.NewClient(ctx, "rollpoker", opt)
	} else {
		client, err = firestore.NewClient(ctx, "rollpoker")
	}
	if err != nil {
		log.Printf("Can't get client: %v", err)
		return
	}
	fmt.Println("Rollpoker started")
}

type GameResponse struct {
	Name string
}

func GenerateNewName() string {
	return "OrangePanda"
}

func MakeTable(w http.ResponseWriter, r *http.Request) {
	var args map[string]string

	err := json.NewDecoder(r.Body).Decode(&args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("args: %v", args)

	var newgame Game
	var settings GameSettings

	settings.GameType = args["GameType"]
	settings.BetLimit = args["BetLimit"]
	i64, _ := strconv.ParseInt(args["StartingChips"], 10, 32)
	settings.StartingChips = int(i64)
	settings.ChipValues = args["ChipValues"] // We don't bother with this just yet, we pass
						 // it straight to javascript.
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

	newgame.Name = "OrangePanda" // GenerateName()
	newgame.Private.PlayerKeys = map[string]string{}
	newgame.Private.TableDecks =	map[string][]string{}
	newgame.Private.AdminPassword = args["AdminPassword"]
	newgame.Private.OrigState = settings

	newgame.Public.Tables = map[string]TableState{}
	newgame.Public.State = NOGAME

	SaveGame(&newgame)

	var newgr GameResponse
	newgr.Name = newgame.Name

	bytes, err := json.Marshal(newgr)
	w.Write(bytes)
}

type GetStateRequest struct {
	Name		string
	Last		int
	PlayerId	string
	PlayerKey	string
}

func FetchGame(name string) *Game {
	var game Game
	gamesRef := client.Doc("games/" + name)
	ctx := context.Background()
	gd, err := gamesRef.Get(ctx)
	if err != nil {
		log.Printf("Can't get snapshot: %v", err)
		return nil
	}
	err = gd.DataTo(&game)
	if err != nil {
		log.Printf("No datato avail: %v", err)
		return nil
	}
	return &game
}

type StateResponse struct {
	Name	string
	GameState	*Game
	Events		[]GameEvent
	Last		int
}

func RegisterAccount(game *Game, gc *GameCommand) bool {
	fmt.Printf("Register: %s %s\n", gc.Args["DisplayName"], gc.Args["Email"])

	if gc.Args["DisplayName"] == "" || gc.Args["Email"] == "" {
		return false
	}
	for _, p := range game.Public.Players {
		if p.DisplayName == gc.Args["DisplayName"] {
			return false
		}
	}

	var player Player

	player.DisplayName = gc.Args["DisplayName"]
	player.PlayerId = GenerateName()
	playerKey := GenerateName()
	game.Private.PlayerKeys[player.PlayerId] = playerKey

	from := mail.NewEmail("RollPoker NoReply", "no-reply@deafcode.com")
	subject := "RollPoker for " + gc.Args["DisplayName"]
	to := mail.NewEmail(gc.Args["DisplayName"], gc.Args["Email"])

	link := "http://localhost/table/" + gc.Name + "?id=" + player.PlayerId + "&key=" + playerKey

	plainTextContent := "You have been invited to join a poker game: " + link
	htmlContent := "<a href=\"" + link + "\">Click here to join the poker game</a>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	sgclient := sendgrid.NewSendClient(SENDGRID_API_KEY)
	_, err := sgclient.Send(message)
	if err != nil {
		return false
	}
	fmt.Println(link)
	if game.Public.Players == nil {
		game.Public.Players = map[string]Player{}
	}
	game.Public.Players[player.PlayerId] = player
	SaveGame(game)

	return err == nil
}

func SaveGame(game *Game) {
	_, err := client.Doc("games/" + game.Name).Set(context.Background(), game)
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	fmt.Println("Saved", game.Name)
}

func Poker(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")
	var gc GameCommand
	var player *Player = nil
	err := json.NewDecoder(r.Body).Decode(&gc)

	if err != nil {
		gc.Name = "OrangePanda"
	}

	game := FetchGame(gc.Name)
	if game == nil {
		http.Error(w, "Unknown Game", http.StatusBadRequest)
		return
	}

	pkey, has := game.Private.PlayerKeys[gc.PlayerId]
	if has && pkey == gc.PlayerKey {
		playerp := game.Public.Players[gc.PlayerId]
		player = &playerp
	}

	fmt.Println("Got", gc.Command)

	if gc.Command == "invite" {
		if !RegisterAccount(game, &gc) {
			http.Error(w, "Unable to register", http.StatusBadRequest)
		}
		return
	}

	if player == nil {
		// If PlayerId is nil, the only thing the player can do is "register" / invite.
		http.Error(w, "You are not a player", http.StatusBadRequest)
		return
	}

	// Call Command by name if it has one
	method := reflect.ValueOf(game).MethodByName(gc.Command)

	if method.IsValid() {
		rval := method.Call([]reflect.Value{reflect.ValueOf(player), reflect.ValueOf(&gc)})
		fmt.Fprintf(w, "%v", rval[0].Bool())
	}
}
