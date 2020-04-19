package rollpoker

import (
	"sort"
	"strconv"
)

// This file contains Hand Ranking methods.
//
// In poker, suit is worthless within a rank, so we consider each card a half-byte.
const (
// 0x075432 is a 7, 5, 4, 3, 2 of any suit, the leading 0 means 'high card'
	HIGHCARD = "0"
// 0x1EE432 is a pair of aces, with 4, 3, 2 as fill
	PAIR = "1"
// 0x2DDCC2 is two pair: Kings and Queens, with deuce as fill
	TWOPAIR = "2"
// 0x3BBB53 is a set of jacks with 5+3 as fill
	SET = "3"
// 0x4A9876 is a ten-high straight: Ten, 9, 8, 7, 6
	STRAIGHT = "4"
// 0x5A9543 is a flush: Ten, 9, 5, 4, 3
	FLUSH = "5"
// 0x6AAADD is a full house: Tens full of aces. (Aces full of tens would be DDDAA)
	FULLHOUSE = "6"
// 0x799996 is 4 nines, with 6 as 5th card
	FOUROFAKIND = "7"
// 0x8A9876 is a straight flush, 10-high
	STRAIGHTFLUSH = "8"
// 0x8EDCBA is a straight flush, ace high ("Royal Flush")
)

// A Texas Hold'em hand has 21 combinations (7*6 / 2*1)
// An Omaha hand has 60, combinations ((5*4)/2)*((4*3/2))

type Card struct {
	Suit string
	Val int // 0x2 (deuce) through 0xE (ace)
}

func EachCombination(cards []*Card, count int, cb func([]*Card)) {
	if len(cards) == count {
		cb(cards)
		return
	}
	for i := 0; i < len(cards); i++ {
		EachCombination(cards[i+1:], count-i, func(rem []*Card) {
			n := []*Card{}
			n = append(n, cards[:i]...)
			cb(append(n,rem...))
		}); }
}

func StrToCard(scard string) *Card {
	var card Card
	card.Suit = scard[0:1]
	switch scard[1:2] {
	case "2": card.Val = 0x2
	case "3": card.Val = 0x3
	case "4": card.Val = 0x4
	case "5": card.Val = 0x5
	case "6": card.Val = 0x6
	case "7": card.Val = 0x7
	case "8": card.Val = 0x8
	case "9": card.Val = 0x9
	case "t": card.Val = 0xa
	case "j": card.Val = 0xb
	case "q": card.Val = 0xc
	case "k": card.Val = 0xd
	case "a": card.Val = 0xe
	}
	return &card
}

func GetTexasRank(hand, board []string) int {
	allcards := make([]*Card,len(hand) + len(board))
	i := 0
	for _, c := range hand {
		allcards[i] = StrToCard(c)
		i++
	}
	for _, c := range board {
		allcards[i] = StrToCard(c)
		i++
	}
	// Sort allcards by high->low Val
	sort.Slice(allcards, func(i,j int) bool {
		// i.Val < j.Val sorts low->high
		return allcards[j].Val < allcards[i].Val
	})
	best := 0
	EachCombination(allcards, 5, func(cards []*Card) {
		val := GetHandVal(cards)
		if val > best { best = val }
	})
	return best
}

func GetHandVal(cards []*Card) int {
	hex := GetHandValS(cards)
	ival, _ := strconv.ParseInt(hex, 16, 32)
	return int(ival)
}

func GetHandValS(cards []*Card) string {
	l := len(cards) // Rather than use 5 everywhere
	hexval := ""
	hasFlush := true
	for i := 1; i < l; i++ {
		if cards[i].Suit != cards[0].Suit { hasFlush = false }
	}

	// Sorted in descending value
	hasStraight := true
	for i := 1; i < l; i++ {
		if cards[i].Val != (cards[i-1].Val - 1) { hasStraight = false }
	}

	if !hasStraight && !hasFlush {
		// Now we count, order by count then by val
		counts := make([]int, 15)
		for i := 0; i < len(counts); i++ { counts[i] = 0; }
		for i := 0; i < l; i++ { counts[cards[i].Val] += 1 }
		sort.Slice(cards, func(i,j int) bool {
			var ic, jc *Card
			ic = cards[i]
			jc = cards[j]
			if counts[jc.Val] < counts[ic.Val] { return true }
			if counts[jc.Val] > counts[ic.Val] { return false }
			// Counts equal:
			return jc.Val < ic.Val
		})
	}
	for i := 0; i < l; i++ {
		hexval = hexval + strconv.FormatInt(int64(cards[i].Val), 16)
	}

	if hasStraight && hasFlush { return STRAIGHTFLUSH + hexval }

	if hasFlush { return FLUSH + hexval }
	if hasStraight { return STRAIGHT + hexval }
	if cards[0].Val == cards[3].Val { return FOUROFAKIND + hexval }

	// Full House?
	if cards[0].Val == cards[2].Val && cards[3].Val == cards[4].Val {
		return FULLHOUSE + hexval
	}

	// Set of 3?
	if cards[0].Val == cards[2].Val {
		return SET + hexval
	}
	// 2 Pair?
	if cards[2].Val == cards[3].Val { return TWOPAIR + hexval }
	// Pair?
	if cards[0].Val == cards[1].Val { return PAIR + hexval }

	return HIGHCARD + hexval
}
