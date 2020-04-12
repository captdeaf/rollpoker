package rollpoker

import (
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
	Table		int	// 0...
	Event		string	// Name for description e.g: "Fold"
	Args		string	// Parameters. "playerid"
	JsonDiff	string	// Diffs of State, TableState, Player and GameSettings
	Log		string	// "Chris folded". If nil, no log.
}

type TableState struct {
	Blinds		[]int
	BlindTime	int
	Pot		int
	Dealer		int
	Deck		[]string `json:"-"`
}

type Game struct {
	State		string
	EventId		int	`json:"-"`
	GameEvents	[]GameEvent	`json:"-"`
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

	newgr.Name = GenerateNewName()

	var newgame Game

	newgame.State = NOGAME
	newgame.Settings = settings
	newgame.CreatedAt = 0
	newgame.UpdatedAt = 0
	newgame.EventId = 1

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

func SendFullState(w http.ResponseWriter, game *Game) {
	// We don't actually send full state.
	// What we do send is:
	// * All Player states (with info stripped out)
	// * All Table states
	// * GameState (running/paused/etc)
	// * GameSettings
	bytes, err := json.Marshal(game)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
	} else {
		fmt.Println("Sending full state of ", game.Settings.GameName)
		w.Write(bytes)
	}
}

func GetState(w http.ResponseWriter, r *http.Request) {
	var gsr GetStateRequest
	err := json.NewDecoder(r.Body).Decode(&gsr)

	if err != nil {
		// http.Error(w, err.Error(), http.StatusBadRequest)
		// return
		gsr.Name = "OrangePanda"
		gsr.Last = 0
	}

	// TODO: Sanity checking on GSR.
	if gsr.PlayerId == "" {
		fmt.Println("Player is nil, hasn't joined")
	}

	game := FetchGame(gsr.Name)
	if game == nil {
		http.Error(w, "Unknown Game", http.StatusBadRequest)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	want := gsr.Last + 1

	if gsr.Last == game.EventId {
		fmt.Fprintf(w, "false")
		return
	}
	// TODO: Sanity Checking

	for _,evt := range game.GameEvents {
		if evt.EventId == want {
			// Only send this Event
			fmt.Fprintf(w, "{}")
			break
		}
	}

	// Send entire state (stripped of other players' personal info, ofc)
	fmt.Println(game.Settings.GameName)
	SendFullState(w, game)
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
	_, err = client.Doc("games/" + gc.Name).Set(context.Background(), game)

	return true
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

	cookie, err := r.Cookie("playerkey")
        if err == nil {
		for _, p := range game.Players {
			if p.PlayerId == gc.PlayerId && p.PlayerKey == cookie.Value {
				player = &p
				break
			}
		}
        }

	if player == nil {
		if gc.Command == "register" {
			if !RegisterAccount(game, &gc) {
				http.Error(w, "Unable to register", http.StatusBadRequest)
			}
		}
		return
	}

	// The only thing the player can do is register.
	fmt.Fprintf(w, "Welcome %s", player.DisplayName)
}


