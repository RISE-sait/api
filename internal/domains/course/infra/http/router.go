package course

import (
	"api/cmd/server/di"
	course "api/internal/domains/course/application"
	"api/internal/domains/course/infra/persistence"

	"github.com/go-chi/chi"
)

func RegisterCourseRoutes(r chi.Router, container *di.Container) {
	coursesHandler := NewHandler(course.NewCourseService(
		&persistence.CourseRepository{
			Queries: container.Queries.CoursesDb,
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
