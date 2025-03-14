package data

import (
	"fmt"
	"math/rand"
)

//
//import (
//	dbSeed "api/cmd/seed/sqlc/generated"
//	"fmt"
//	"math/rand"
//)
//
//func GetGames() {
//
//	for i := 0; i < 20; i++ {
//		name := generateCourseName(i)
//		description := generateCourseDescription(name)
//		capacity := generateCourseCapacity()
//	}
//
//	return dbSeed.InsertCoursesParams{
//		NameArray:        nameArray,
//		DescriptionArray: descriptionArray,
//		CapacityArray:    capacityArray,
//	}
//}

func GenerateGameName(index int) string {
	gameTypes := []string{"Soccer", "Basketball", "Chess", "Tennis", "Hockey", "Baseball"}
	formats := []string{"Tournament", "League", "Match", "Championship", "Cup"}
	return fmt.Sprintf("%s %s %d", gameTypes[rand.Intn(len(gameTypes))], formats[rand.Intn(len(formats))], index+1)
}
