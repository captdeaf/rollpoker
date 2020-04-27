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

RegisterTest("Decrypt Hand Vals:", func() bool {
	val := "hasa" // ace of hearts, ace of spades
	dec := rollpoker.DecryptHand(val, "abc123")
	return dec == val
})

RegisterTest("StrHand To Hand Vals:", func() bool {
	ret := rollpoker.HandToHandVals("hasa")
	if ret[0] != "ha" { return false }
	if ret[1] != "sa" { return false }
	return true
})

}
