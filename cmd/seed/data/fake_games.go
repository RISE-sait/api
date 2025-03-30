package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"fmt"
	"github.com/google/uuid"
	"math/rand"
)

func GetGames(numGames int, teamIds []uuid.UUID) dbSeed.InsertGamesParams {
	params := dbSeed.InsertGamesParams{
		NameArray:        make([]string, numGames),
		DescriptionArray: make([]string, numGames),
		LevelArray:       make([]dbSeed.ProgramProgramLevel, numGames),
		WinTeamArray:     make([]uuid.UUID, numGames),
		LoseTeamArray:    make([]uuid.UUID, numGames),
		WinScoreArray:    make([]int32, numGames),
		LoseScoreArray:   make([]int32, numGames),
	}

	for i := 0; i < numGames; i++ {
		params.NameArray[i] = generateGameName(i)
		params.DescriptionArray[i] = generateGameDescription(i)
		params.LevelArray[i] = dbSeed.ProgramProgramLevelAll
		params.WinTeamArray[i] = teamIds[i%len(teamIds)]
		params.LoseTeamArray[i] = teamIds[(i+1)%len(teamIds)]
		params.WinScoreArray[i] = int32(21 + i%15)
		params.LoseScoreArray[i] = int32(15 + i%10)
	}

	return params
}

func generateGameName(index int) string {
	gameTypes := []string{"Soccer", "Basketball", "Hockey"}
	formats := []string{"Tournament", "League", "Match"}
	return fmt.Sprintf("%s %s %d", gameTypes[rand.Intn(len(gameTypes))], formats[rand.Intn(len(formats))], index+1)
}

func generateGameDescription(index int) string {
	return fmt.Sprintf("This is a description for game %d", index+1)
}
