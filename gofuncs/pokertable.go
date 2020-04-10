package rollpoker

import (
	"os"
	"fmt"
	"context"
	"log"
	"net/http"
	// "encoding/json"

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
	State		string
	Hand		string
	Bet		int // Current amount bet // not in the pot.
}

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
	BlindTime	uint64
	Pot		int
	Deck		[]string
	GameEvents	[]GameEvent
	GameLogs	[]GameLog
}

type Game struct {
	Players		map[int] Player
	State		GameState
	Settings	GameSettings
	Body		string
	// Timestamps for clearing from database
	CreatedAt	uint64
	UpdatedAt	uint64
}

type GameCommand struct {
	Command		string `json:"command"`
	Args		string `json:"args"`
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
		log.Fatalf("Can't get client: %v", err)
		return
	}
	fmt.Println("Rollpoker started")
}

type Foo struct {
	Name string
}

func MakeTable(w http.ResponseWriter, r *http.Request) {
	gamesRef := client.Doc("games/OrangeShipwreck")
	ctx := context.Background()
	fmt.Println(gamesRef)
	gd, err := gamesRef.Get(ctx)
	fmt.Println(gd)
	fmt.Println(gd.Data())
	if err != nil {
		log.Fatalf("Can't get snapshot: %v", err)
		return
	}
	var g *Foo
	err = gd.DataTo(&g)
	if err != nil {
		log.Fatalf("No datato avail: %v", err)
		return
	}

	// TODO: Sanity Checking
	fmt.Fprintf(w, "Welcome %s", g.Name)

	/*
	err = json.NewDecoder(r.Body).Decode(&settings)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	*/
}
