// Test the hand encryption
package main

import (
	"deafcode.com/rollpoker"
)


func init() {

RegisterTest("Encryption:", func() bool {
	key := "FooBarBazFizbit"
	val := "hasa" // ace of hearts, ace of spades
	enc := rollpoker.EncryptHand(val, key)
	dec := rollpoker.DecryptHand(enc, key)
	return dec == val
})

}
