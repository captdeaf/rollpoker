package rollpoker

import (
	"os"
	"context"
	"fmt"
	"net/http"

	"google.golang.org/api/option"
	firebase "firebase.google.com/go"
	firebaseauth "firebase.google.com/go/auth"
)

var FBAUTH_CLIENT *firebaseauth.Client

func GetUserIDFromHeader(req *http.Request) string {
	// 1) Extract the token from Authorization: Bearer
	header := req.Header.Get("Authorization")
	if header == "" { return "" }
	if header[0:7] != "Bearer " { return "" }
	jwtToken := header[7:]

	ctx := context.Background()

	// 2) Init FBAUTH_CLIENT if it's not already
	if FBAUTH_CLIENT == nil {
		_, oserr := os.Stat("firebase-key.json")
		var clienterr error
		if oserr == nil {
			// Localhost
			opt := option.WithCredentialsFile("firebase-key.json")
			fbapp, apperr := firebase.NewApp(ctx, nil, opt)
			if apperr != nil {
				panic(fmt.Sprintf("Unable to get NewApp: %v\n", apperr))
				return ""
			}
			FBAUTH_CLIENT, clienterr = fbapp.Auth(ctx)
		} else {
			// Server
			fbapp, apperr := firebase.NewApp(ctx, nil)
			if apperr != nil {
				panic(fmt.Sprintf("Unable to get NewApp: %v\n", apperr))
				return ""
			}
			FBAUTH_CLIENT, clienterr = fbapp.Auth(ctx)
		}
		if clienterr != nil {
			panic(fmt.Sprintf("Unable to get FBA Client: %v\n", clienterr))
			return ""
		}
	}

	// 3) Verify the token
	firebaseToken, err := FBAUTH_CLIENT.VerifyIDTokenAndCheckRevoked(ctx, jwtToken)

	if err != nil {
		fmt.Errorf("Unable to verify Token: %v\n", err)
		return ""
	}
	return firebaseToken.UID
}
