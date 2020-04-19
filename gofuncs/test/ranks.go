// Test the hand ranker
package main

import (
	"strings"

	"deafcode.com/rollpoker"
)


func Rank(cardstr string) int {
	cards := strings.Fields(cardstr)
	return rollpoker.GetTexasRank(cards, []string{})
}

func init() {

RegisterTest("0x000000 < High Card < 0x100000", func() bool {
	return Rank("ha st d3 c8 d9") < 0x100000
})
RegisterTest("0x100000 < Pair      < 0x200000", func() bool {
	val := Rank("ha sa d3 c8 d9")
	if val < 0x100000 { return false }
	return val < 0x200000
})
RegisterTest("0x200000 < Two Pair  < 0x300000", func() bool {
	val := Rank("ha sa d3 c3 d9")
	if val < 0x200000 { return false }
	return val < 0x300000
})
RegisterTest("0x300000 < Set       < 0x400000", func() bool {
	val := Rank("ha sa da c8 d9")
	if val < 0x300000 { return false }
	return val < 0x400000
})
RegisterTest("Rank Set > Two Pair", func() bool {
	return Rank("ha sa da c8 d9") > Rank("ha sa ht dt c9")
})

RegisterTest("Ten High > Nine High", func() bool {
	return Rank("ht s9 d4 c5 d7") > Rank("h6 s9 h3 d7 c8")
})

RegisterTest("Nine > Eight High", func() bool {
	return Rank("h8 s3 d4 c5 d7") < Rank("h6 s9 h3 d7 c8")
})

RegisterTest("Nine high flush > Eight High", func() bool {
	return Rank("h8 h3 h4 h5 h7") < Rank("h6 h9 h3 h7 h8")
})

RegisterTest("Flush found in 7 cards", func() bool {
	return Rank("h9 sa h3 st h2 h7 h8") == Rank("h9 h3 h2 h7 h8")
})

RegisterTest("Higher full house found in 7 cards", func() bool {
	return Rank("h9 s9 h3 s3 c3 d9 h8") == Rank("h9 s9 d9 c3 s3")
})

RegisterTest("Straight w/ different suits identical", func() bool {
	return Rank("s3 s4 s5 c6 d7") == Rank("d3 d4 s5 c6 d7")
})

}
