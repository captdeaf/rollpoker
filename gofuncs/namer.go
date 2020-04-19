package rollpoker

import (
	"fmt"
	"math/rand"
)

func GenerateName() string {
	adj := ALL_ADJECTIVES[rand.Intn(len(ALL_ADJECTIVES))]
	col := ALL_COLORS[rand.Intn(len(ALL_COLORS))]
	noun := ALL_NOUNS[rand.Intn(len(ALL_NOUNS))]
	return fmt.Sprintf("%s%s%s", adj, col, noun)
}
