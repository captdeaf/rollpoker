package rollpoker

import (
	"reflect"
	"os"
	"fmt"
	"context"
	"log"
	"net/http"
	"encoding/json"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type GameSettings struct {
	GameName	string // TexasHoldem, OmahaHoldem, etc.
	GameType	string // Cash, SitNGo
	BetLimit	string // NoLimit, PotLimit
	StartingChips	string // 1500. (Yes, string in case of other args)
	StartBlinds	string // "25 50"
	ChipValues	string // White Red Blue Green Black Yellow: "25 100 500 1000..."
	BlindStructure	string // 25 50,25 75,50 100,75 150,...
	BlindTimes	string // 40 40 40 20, for first 3 to be 40 minutes, then 20 mins after that.
	AdminPassword	string `json:"-"` // "hunter2"
}

type Player struct {
	PlayerId	string	// "p3253c321"
	Table		int	// 0...?
	Seat		int	// 0...9
	DisplayName	string	// What to show everyone else
	PlayerKey	string	`json:"-"` // To reconnect / view from multiple devices
	Chips		int	// How many chips they have

	// Current hand info:
	State		string
	Hand		string	`json:"-"` // To reconnect / view from multiple devices
	Bet		int	// Current amount bet // not in the pot.
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

type GameEvent struct {
	EventId		int	// 0..9
	PlayerId	string	// MadOrangeCow
	Event		string	// Name for description e.g: "Fold"
	Args		string	// Parameters. "playerid"
	Log		string	// "Chris folded". If nil, no log.
}

func AddEvent(game *Game, event string, args interface{}, logmsg string) {
	evt := GameEvent{}
	game.EventId += 1
	evt.EventId = game.EventId
	evt.Log = logmsg
	evt.Event = event
	bytes, err := json.Marshal(args)
	if err != nil {
		fmt.Printf("Error with event '%s': %v", event, err)
	} else {
		evt.Args = string(bytes)
	}
}

type TableState struct {
	Blinds		[]int
	BlindTime	int
	Pot		int
	Dealer		int
	Deck		[]string `json:"-"`
}

type Game struct {
	Name		string
	State		string
	EventId		int	`json:"-"`
	Events	[]GameEvent	`json:"-"`
	Players		map[string] Player
	TableStates	[]TableState
	Settings	GameSettings
	CreatedAt	int	`json:"-"`
	UpdatedAt	int	`json:"-"`
}

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
	var settings GameSettings

	err := json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var newgr GameResponse

	var newgame Game

	newgame.Name = GenerateName()
	newgame.State = NOGAME
	newgame.Settings = settings
	newgame.CreatedAt = 0
	newgame.UpdatedAt = 0
	newgame.EventId = -1

	newgr.Name = newgame.Name

	_, err = client.Doc("games/" + newgr.Name).Set(context.Background(), newgame)

	if err != nil {
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

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
	GameName	string
	GameState	*Game
	Events		[]GameEvent
	Last		int
}

func GetState(w http.ResponseWriter, r *http.Request) {
	var gsr GetStateRequest
	err := json.NewDecoder(r.Body).Decode(&gsr)

	if err != nil {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		// return
		gsr.Name = "OrangePanda"
		gsr.Last = -1
	}

	// TODO: Sanity checking on GSR.
	if gsr.PlayerId == "" {
		fmt.Println("Player is nil, hasn't joined")
	} else {
		fmt.Println("Player is", gsr.PlayerId)
	}

	game := FetchGame(gsr.Name)
	if game == nil {
		http.Error(w, "Unknown Game", http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")

	if gsr.Last == game.EventId {
		fmt.Fprintf(w, "false")
		return
	}

	resp := StateResponse{}
	resp.GameName = game.Name
	resp.GameState = game
	if gsr.Last >= 0 && game.EventId >= 0 {
		resp.Events = game.Events[gsr.Last:game.EventId]
	}
	resp.Last = game.EventId

	// Send entire state (stripped of other players' personal info, ofc)
	fmt.Println(game.Settings.GameName)

	bytes, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		fmt.Println("Sending full state of ", game.Settings.GameName)
		w.Write(bytes)
	}
}

func RegisterAccount(game *Game, gc *GameCommand) bool {
	fmt.Printf("Register: %s %s\n", gc.Args["DisplayName"], gc.Args["Email"])

	if gc.Args["DisplayName"] == "" || gc.Args["Email"] == "" {
		return false
	}
	for _, p := range game.Players {
		if p.DisplayName == gc.Args["DisplayName"] {
			return false
		}
	}

	var player Player

	player.DisplayName = gc.Args["DisplayName"]
	player.Table = -1
	player.Seat = -1
	player.PlayerId = GenerateName()
	player.PlayerKey = GenerateName()
	player.Chips = -1

	player.State = "BUSTED"
	player.Hand = ""
	player.Bet = -1

	from := mail.NewEmail("RollPoker NoReply", "no-reply@deafcode.com")
	subject := "RollPoker for " + gc.Args["DisplayName"]
	to := mail.NewEmail(gc.Args["DisplayName"], gc.Args["Email"])

	link := "http://localhost/table/" + gc.Name + "?id=" + player.PlayerId + "&key=" + player.PlayerKey

	plainTextContent := "You have been invited to join a poker game: " + link
	htmlContent := "<a href=\"" + link + "\">Click here to join the poker game</a>"
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, htmlContent)
	sgclient := sendgrid.NewSendClient(SENDGRID_API_KEY)
	_, err := sgclient.Send(message)
	if err != nil {
		return false
	}
	fmt.Println(link)
	if game.Players == nil {
		game.Players = map[string]Player{}
	}
	game.Players[player.PlayerId] = player
	SaveGame(game)

	return err == nil
}

func SaveGame(game *Game) {
	fmt.Println("Saved", game.Name)
	client.Doc("games/" + game.Name).Set(context.Background(), game)
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

	for _, p := range game.Players {
		if p.PlayerId == gc.PlayerId && p.PlayerKey == gc.PlayerKey {
			player = &p
			break
		}
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
		return;
	}

	fmt.Fprintf(w, "Welcome %s", player.DisplayName)
	// Call Command by name if it has one
	method := reflect.ValueOf(game).MethodByName(gc.Command)

	if method.IsValid() {
		method.Call([]reflect.Value{reflect.ValueOf(player), reflect.ValueOf(&gc)})
	}
}

func (game *Game) StartGame(player *Player, gc *GameCommand) {
	fmt.Printf("Player %s starts game %s with command %s",
			player.DisplayName, game.Name, gc.Command)
}
