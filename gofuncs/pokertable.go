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
}

type Player struct {
	Email		string
	Passphrase	string
	DisplayName	string
	Chips		int

	// Current hand info:
	State		int
	Hand		string
	Bet		int // Current amount bet // not in the pot.
}

const (
	BUSTED = iota // Never changes unless restart or cash buyin.
	FOLDED	// Out of this hand

	WAITING // Hasn't had their turn yet.
	BET	// Bet or raised
	CALLED	// All players but one must be CALLED (or ALLIN)
	ALLIN   // Has no more chips left.
)

type GameEvent struct {
	EventId		int
	Event		string
	Args		string
}

type GameLog struct {
	Player		string // seat0 seat1
	Message		string
}

type GameState struct {
	Running		bool
	Blinds		[]int
	BlindTime	int
	Pot		int
	Deck		[]string
	GameEvents	[]GameEvent
	GameLogs	[]GameLog
}

type Game struct {
	Players		map[string] Player
	State		GameState
	Settings	GameSettings
	CreatedAt	int
	UpdatedAt	int
}

type GameCommand struct {
	Command		string `json:"command"`
	Args		string `json:"args"`
}

var client *firestore.Client

func init() {
	fmt.Printf("Waiting %d and Busted %d", WAITING, BUSTED)
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
		log.Fatalf("Can't get client: %v", err)
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

	newgame.State.Running = false
	newgame.State.BlindTime = 0
	newgame.State.Pot = 0
	newgame.Settings = settings
	newgame.CreatedAt = 0
	newgame.UpdatedAt = 0

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
	Player		string
	PlayerPass	string
}

func FetchGame(name string) *Game {
	var game Game
	gamesRef := client.Doc("games/" + name)
	ctx := context.Background()
	gd, err := gamesRef.Get(ctx)
	fmt.Println(gd.Data())
	if err != nil {
		log.Fatalf("Can't get snapshot: %v", err)
		return nil
	}
	err = gd.DataTo(&game)
	if err != nil {
		log.Fatalf("No datato avail: %v", err)
		return nil
	}
	return &game
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
	if gsr.Player == "" {
		fmt.Println("Player is nil, hasn't joined")
	}

	game := FetchGame(gsr.Name)
	if game == nil {
		http.Error(w, "Unknown Game", http.StatusBadRequest)
		return
	}


	// TODO: Sanity Checking
	fmt.Println(game.Settings.GameName)
	fmt.Fprintf(w, "{}")
}

func Poker(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "{}")
}


