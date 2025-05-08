package data

import (
	"math/rand"
	"time"

	dbSeed "api/cmd/seed/sqlc/generated"

	"github.com/google/uuid"
)

// GetGames generates a batch of fake game data for seeding the database.
// It returns a populated InsertGamesParams struct with randomized teams, scores, times, and statuses.
func GetGames(teamIDs []uuid.UUID, locationNames []string, numGames int) dbSeed.InsertGamesParams {
	var (
		startTimes        []time.Time // Game start times
		endTimes          []time.Time // Game end times
		homeTeamIDs       []uuid.UUID // Home team UUIDs
		awayTeamIDs       []uuid.UUID // Away team UUIDs
		locationNameArray []string    // Location names
		homeScores        []int32     // Home team scores
		awayScores        []int32     // Away team scores
		statuses          []string    // Game status values
	)

	for i := 0; i < numGames; i++ {
		// Randomly select different home and away teams
		home := teamIDs[rand.Intn(len(teamIDs))]
		away := teamIDs[rand.Intn(len(teamIDs))]
		for home == away {
			away = teamIDs[rand.Intn(len(teamIDs))]
		}

		// Set start time to i days from now, and end time 90 minutes later
		start := time.Now().Add(time.Duration(i) * 24 * time.Hour)
		end := start.Add(90 * time.Minute)

		// Build up all the slices to return
		homeTeamIDs = append(homeTeamIDs, home)
		awayTeamIDs = append(awayTeamIDs, away)
		locationNameArray = append(locationNameArray, locationNames[rand.Intn(len(locationNames))])
		homeScores = append(homeScores, int32(rand.Intn(5))) // Random score: 0–4
		awayScores = append(awayScores, int32(rand.Intn(5))) // Random score: 0–4
		statuses = append(statuses, []string{"scheduled", "completed", "canceled"}[rand.Intn(3)])
		startTimes = append(startTimes, start)
		endTimes = append(endTimes, end)
	}

	// Return the fully populated seed parameter object
	return dbSeed.InsertGamesParams{
		HomeTeamIds:   homeTeamIDs,
		AwayTeamIds:   awayTeamIDs,
		LocationNames: locationNameArray,
		HomeScores:    homeScores,
		AwayScores:    awayScores,
		StartTimes:    startTimes,
		EndTimes:      endTimes,
		Statuses:      statuses,
	}
}
