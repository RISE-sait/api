package mapper

import (
	entity "api/internal/domains/course/entities"
	"api/internal/domains/course/infra/http/dto"
)

func MapEntityToResponse(course *entity.Course) dto.CourseResponse {
	return dto.CourseResponse{
		ID:        course.ID,
		Name:      course.Name,
		StartDate: course.StartDate,
		EndDate:   course.EndDate,
	}
}
