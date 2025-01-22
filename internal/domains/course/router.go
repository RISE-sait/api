package course

import (
	"api/internal/domains/course/infra"
	db "api/internal/domains/course/infra/sqlc"
	"github.com/go-chi/chi"
)

func RegisterCourseRoutes(r chi.Router, queries *db.Queries) {
	coursesHandler := NewHandler(NewService(
		&course.Repository{
			Queries: queries,
		},
	))

	r.Route("/courses", func(auth chi.Router) {
		auth.Get("/", coursesHandler.GetAllCourses)
		auth.Get("/{id}", coursesHandler.GetCourseById)
		auth.Post("/", coursesHandler.CreateCourse)
		auth.Put("/{id}", coursesHandler.UpdateCourse)
		auth.Delete("/{id}", coursesHandler.DeleteCourse)
	})
}
