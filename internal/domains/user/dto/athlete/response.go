package athlete

import (
	values "api/internal/domains/user/values"
)

type ResponseAthlete struct {
	ID        string  `json:"id"`
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Points    int32   `json:"points"`
	Wins      int32   `json:"wins"`
	Losses    int32   `json:"losses"`
	Assists   int32   `json:"assists"`
	Rebounds  int32   `json:"rebounds"`
	Steals    int32   `json:"steals"`
	PhotoURL  *string `json:"photo_url"`
	TeamID    *string `json:"team_id"`
}

func FromReadValue(v values.AthleteReadValue) ResponseAthlete {
	var teamID *string
	if v.TeamID != nil {
		t := v.TeamID.String()
		teamID = &t
	}

	return ResponseAthlete{
		ID:        v.ID.String(),
		FirstName: v.FirstName,
		LastName:  v.LastName,
		Points:    v.Points,
		Wins:      v.Wins,
		Losses:    v.Losses,
		Assists:   v.Assists,
		Rebounds:  v.Rebounds,
		Steals:    v.Steals,
		PhotoURL:  v.PhotoURL,
		TeamID:    teamID,
	}
}
