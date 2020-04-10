// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package gopher contains an HTTP function that shows a gopher.
package rollpoker

import (
	"fmt"
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"

)

type Game struct {
	Name		string `firestore:"name" json:"name"`
}

var client *firestore.Client

func init() {
	ctx := context.Background()
	opt := option.WithCredentialsFile("firebase-key.json")
	var err error
	client, err = firestore.NewClient(ctx, "rollpoker", opt)
	if err != nil {
		log.Fatalf("Can't get client: %v", err)
		return
	}
	fmt.Println("Rollpoker started")

	gamesRef := client.Doc("games/OrangeShipwreck")
	fmt.Println(gamesRef)
	gd, err := gamesRef.Get(ctx)
	fmt.Println(gd)
	fmt.Println(gd.Data())
	if err != nil {
		log.Fatalf("Can't get snapshot: %v", err)
		return
	}
	var g *Game
	err = gd.DataTo(&g)
	if err != nil {
		log.Fatalf("No datato avail: %v", err)
		return
	}
	fmt.Println("Got: ",g.Name)
}

func MakeTable(w http.ResponseWriter, r *http.Request) {
}
