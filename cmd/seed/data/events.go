package data

import (
	dbSeed "api/cmd/seed/sqlc/generated"
	"api/internal/custom_types"
	"api/internal/libs/validators"
	"log"
	"strings"
	"time"
)

func GetEvents() dbSeed.InsertEventsParams {

	var (
		recurrenceStartAtArray []time.Time
		recurrenceEndAtArray   []time.Time
		eventStartTimeArray    []custom_types.TimeWithTimeZone
		eventEndTimeArray      []custom_types.TimeWithTimeZone
		dayArray               []dbSeed.DayEnum
		programNameArray       []string
		//courseNameArray       []string
		//gameNameArray         []string
		locationNameArray []string
	)

	for _, practice := range Practices {

		for _, schedule := range practice.Schedules {

			var programStartAt time.Time
			if schedule.ProgramStartDate != "" {
				parsedStart, err := time.Parse("2006-01-02", schedule.ProgramStartDate)
				if err == nil {
					programStartAt = parsedStart
				} else {
					log.Fatalf("Error parsing program start date: %v", err)
				}
			} else {
				programStartAt = time.Now() // Default to today if empty
			}

			// Handle ProgramEndDate
			var programEndAt time.Time

			if schedule.ProgramEndDate != "" {

				parsedEnd, err := time.Parse("2006-01-02", schedule.ProgramEndDate)
				if err != nil {
					log.Fatalf("Error parsing program end date: %v", err)
				}
				if programStartAt.Year() == parsedEnd.Year() &&
					programStartAt.Month() == parsedEnd.Month() &&
					programStartAt.Day() == parsedEnd.Day() {

					// Set start to 00:01 and end to 23:59 on the same day
					programStartAt = time.Date(
						programStartAt.Year(),
						programStartAt.Month(),
						programStartAt.Day(),
						0, 1, 0, 0, // 00:01
						programStartAt.Location(),
					)

					parsedEnd = time.Date(
						parsedEnd.Year(),
						parsedEnd.Month(),
						parsedEnd.Day(),
						23, 59, 0, 0, // 23:59
						parsedEnd.Location(),
					)
				}

				programEndAt = parsedEnd
			} else {

				programEndAt = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC) // Default to zero time if empty
			}

			recurrenceStartAtArray = append(recurrenceStartAtArray, programStartAt)
			recurrenceEndAtArray = append(recurrenceEndAtArray, programEndAt)

			eventStartTime, err := validators.ParseTime(schedule.EventStartTime + ":00+00:00")

			if err != nil {
				log.Fatalf("Failed to parse session start time: %v", err)
				return dbSeed.InsertEventsParams{}
			}

			eventEndTime, err := validators.ParseTime(schedule.EventEndTime + ":00+00:00")

			if err != nil {
				log.Fatalf("Failed to parse session end time: %v", err)
				return dbSeed.InsertEventsParams{}
			}

			eventStartTimeArray = append(eventStartTimeArray, eventStartTime)
			eventEndTimeArray = append(eventEndTimeArray, eventEndTime)

			locationNameArray = append(locationNameArray, schedule.Location)

			day := dbSeed.DayEnum(strings.ToUpper(schedule.Day))

			if !day.Valid() {
				log.Fatalf("Invalid day: %v", schedule.Day)
				return dbSeed.InsertEventsParams{}
			}

			dayArray = append(dayArray, day)

			programNameArray = append(programNameArray, practice.Name)

		}
	}

	return dbSeed.InsertEventsParams{
		RecurringStartAtArray: recurrenceStartAtArray,
		RecurringEndAtArray:   recurrenceEndAtArray,
		EventStartTimeArray:   eventStartTimeArray,
		EventEndTimeArray:     eventEndTimeArray,
		DayArray:              dayArray,
		ProgramNameArray:      programNameArray,
		LocationNameArray:     locationNameArray,
	}
}
