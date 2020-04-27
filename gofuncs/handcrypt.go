package rollpoker

import (
	"fmt"
	"encoding/base64"
	"math/rand"
)

var ALL_CHARS = []byte("0123456789abcdefghmijklnopqrstuvwxyz")
var MID_BYTE = byte('m')

func EncryptHand(hand, key string) string {
	splat := make([]byte, len(ALL_CHARS))
	copy(splat, ALL_CHARS)
	rand.Shuffle(len(splat), func(i,j int) { splat[i], splat[j] = splat[j], splat[i] })
	bkey := []byte(key)
	bhand := []byte(hand)

	result := make([]byte, len(splat) + len(hand) + 1)
	i := 0
	for idx := 0; idx < len(splat); idx++ {
		result[i] = splat[idx]
		i++
		if splat[idx] == MID_BYTE {
			copy(result[i:i+len(bhand)], bhand)
			i += len(bhand)
			result[i] = MID_BYTE
			i++
		}
	}

	l := len(bkey)
	for j := 0; j < len(result); j++ {
		result[j] = result[j] ^ bkey[j%l]
	}
	return "!" + base64.StdEncoding.EncodeToString(result)
}

func DecryptHand(msg, key string) string {
	if len(msg) == 0 { return "" } // Empty string
	if msg[0:1] != "!" { return msg } // It's already decrypted
	bmsg, err := base64.StdEncoding.DecodeString(msg[1:])

	if err != nil {
		fmt.Println("decode error:", err)
		return "Undecipherable"
	}
	bkey := []byte(key)

	result := make([]byte, len(msg))
	l := len(bkey)
	for j := 0; j < len(bmsg); j++ {
		result[j] = bmsg[j] ^ bkey[j%l]
	}

	for s := 0; s < len(result); s++ {
		if result[s] == MID_BYTE {
			s++
			for e := s ; e < len(result); e++ {
				if result[e] == MID_BYTE {
					return string(result[s:e])
				}
			}
		}
	}
	fmt.Printf("ERROR: Undecipherable! '%v' and '%v'\n", msg, string(result))
	return string("Undecipherable")
}

func HandToHandVals(strhand string) []string {
	hand := make([]string, len(strhand)/2)

	for i := 0; i < len(strhand); i += 2 {
		hand[i/2] = strhand[i:i+2]
	}
	return hand
}

func GetHandVals(game *Game, player *Player) []string {
	pkey := game.Private.PlayerKeys[player.PlayerId]
	res := DecryptHand(player.Hand, pkey)
	return HandToHandVals(res)
}
