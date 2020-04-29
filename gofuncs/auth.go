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
		var err error
		_, err = os.Stat("firebase-key.json")
		if err == nil {
			opt := option.WithCredentialsFile("firebase-key.json")
			fbapp, err := firebase.NewApp(ctx, nil, opt)
			if err != nil { return "" }
			FBAUTH_CLIENT, err = fbapp.Auth(ctx)
		} else {
			fbapp, err := firebase.NewApp(ctx, nil)
			if err != nil { return "" }
			FBAUTH_CLIENT, err = fbapp.Auth(ctx)
		}
		if err != nil || FBAUTH_CLIENT == nil {
			err = fmt.Errorf("error getting Auth client: %v", err)
			return ""
		}
	}

	// 3) Verify the token
	firebaseToken, err := FBAUTH_CLIENT.VerifyIDTokenAndCheckRevoked(ctx, jwtToken)

	if err != nil {
		fmt.Errorf("Error w/ token: %v\n", err)
		return ""
	}
	return firebaseToken.UID
}
