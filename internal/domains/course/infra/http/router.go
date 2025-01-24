package course

import (
	course "api/internal/domains/course/application"
	"api/internal/domains/course/infra/persistence"
	db "api/internal/domains/course/infra/persistence/sqlc/generated"

	"github.com/go-chi/chi"
)

func RegisterCourseRoutes(r chi.Router, queries *db.Queries) {
	coursesHandler := NewHandler(course.NewCourseService(
		&persistence.CourseRepository{
			Queries: queries,
		},
	))

	r.Route("/courses", func(auth chi.Router) {
		auth.Get("/", coursesHandler.GetAllCourses)
		auth.Get("/{id}", coursesHandler.GetCourseById)
		auth.Post("/", coursesHandler.CreateCourse)
		auth.Put("/", coursesHandler.UpdateCourse)
		auth.Delete("/{id}", coursesHandler.DeleteCourse)
	})
}
