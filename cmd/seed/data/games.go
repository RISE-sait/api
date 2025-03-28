package data

import (
	"fmt"
	"math/rand"
)

func GenerateGameName(index int) string {
	gameTypes := []string{"Soccer", "Basketball", "Hockey"}
	formats := []string{"Tournament", "League", "Match"}
	return fmt.Sprintf("%s %s %d", gameTypes[rand.Intn(len(gameTypes))], formats[rand.Intn(len(formats))], index+1)
}

func GenerateGameDescription(index int) string {
	return fmt.Sprintf("This is a description for game %d", index+1)
}
