package rollpoker

import (
	"fmt"
	"math/rand"
	"time"
)

func GenerateName() string {
	rand.Seed(time.Now().UnixNano())
	adj := ALL_ADJECTIVES[rand.Intn(len(ALL_ADJECTIVES))]
	col := ALL_COLORS[rand.Intn(len(ALL_COLORS))]
	noun := ALL_NOUNS[rand.Intn(len(ALL_NOUNS))]
	return fmt.Sprintf("%s%s%s", adj, col, noun)
}
