package mapper

import (
	entity "api/internal/domains/course/entities"
	"api/internal/domains/course/infra/http/dto"
)

func MapCreateRequestToEntity(body dto.CreateCourseRequestBody) entity.Course {
	return entity.Course{
		Name:        body.Name,
		Description: body.Description,
		StartDate:   body.StartDate,
		EndDate:     body.EndDate,
	}
}

func MapUpdateRequestToEntity(body dto.UpdateCourseRequest) entity.Course {
	return entity.Course{
		ID:        body.ID,
		Name:      body.Name,
		StartDate: body.StartDate,
		EndDate:   body.EndDate,
	}
}

func MapEntityToResponse(course *entity.Course) dto.CourseResponse {
	return dto.CourseResponse{
		ID:        course.ID,
		Name:      course.Name,
		StartDate: course.StartDate,
		EndDate:   course.EndDate,
	}
}
