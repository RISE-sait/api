package router

import (
	_interface "api/cmd/server/interface"
	"api/internal/domains/course"
	"api/internal/domains/identity"

	"github.com/go-chi/chi"
)

func RegisterRoutes(r *chi.Mux, queries _interface.QueriesType) {

	identity.RegisterIdentityRoutes(r, queries.IdentityDb)
	course.RegisterCourseRoutes(r, queries.CoursesDb)
}
