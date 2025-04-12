package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"fmt"
	"math/rand"
)

func generateCourseName(index int) string {
	adjectives := []string{"Advanced", "Beginner", "Professional", "Essential", "Expert"}
	subjects := []string{"Python", "Data Science", "Machine Learning", "Cybersecurity", "Web Development", "Cloud Computing"}
	return fmt.Sprintf("%s %s %d", adjectives[rand.Intn(len(adjectives))], subjects[rand.Intn(len(subjects))], index+1)
}

// Generate random course description
func generateCourseDescription(name string) string {
	templates := []string{
		"This course provides an in-depth understanding of %s.",
		"Master the fundamentals of %s with hands-on projects.",
		"Learn %s from industry experts through interactive lessons.",
		"An intensive course on %s, covering all essential topics.",
		"Gain practical skills in %s with real-world applications.",
	}
	return fmt.Sprintf(templates[rand.Intn(len(templates))], name)
}

func GetCourses() dbSeed.InsertCoursesParams {

	var (
		nameArray        []string
		descriptionArray []string
		levelArray       []dbSeed.ProgramProgramLevel
		isPayPerEvent    []bool
	)

	for i := 0; i < 20; i++ {
		name := generateCourseName(i)
		description := generateCourseDescription(name)

		nameArray = append(nameArray, name)
		descriptionArray = append(descriptionArray, description)
		levelArray = append(levelArray, dbSeed.ProgramProgramLevelIntermediate)
		isPayPerEvent = append(isPayPerEvent, true)
	}

	return dbSeed.InsertCoursesParams{
		NameArray:          nameArray,
		DescriptionArray:   descriptionArray,
		LevelArray:         levelArray,
		IsPayPerEventArray: isPayPerEvent,
	}
}
