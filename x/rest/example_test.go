package rest_test

import (
	"context"
	"fmt"

	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/x/rest"
)

type exampleUser struct {
	ID string
}

type exampleUserRepo struct{}

func (exampleUserRepo) FindAll(_ context.Context, _ *rest.QueryParams) ([]exampleUser, int64, error) {
	return nil, 0, nil
}

func (exampleUserRepo) FindByID(_ context.Context, _ string) (*exampleUser, error) {
	return nil, nil
}

func (exampleUserRepo) Create(_ context.Context, _ *exampleUser) error {
	return nil
}

func (exampleUserRepo) Update(_ context.Context, _ string, _ *exampleUser) error {
	return nil
}

func (exampleUserRepo) Delete(_ context.Context, _ string) error {
	return nil
}

func (exampleUserRepo) Count(_ context.Context, _ *rest.QueryParams) (int64, error) {
	return 0, nil
}

func (exampleUserRepo) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func ExampleRegisterResourceRoutes() {
	r := router.NewRouter()
	spec := rest.DefaultResourceSpec("users").WithPrefix("/api/users")
	controller := rest.NewDBResource[exampleUser](spec, exampleUserRepo{})

	rest.RegisterResourceRoutes(r, spec.Prefix, controller, rest.DefaultRouteOptions())

	for _, route := range r.Routes() {
		fmt.Printf("%s %s\n", route.Method, route.Path)
	}

	// Output:
	// DELETE /api/users/:id
	// DELETE /api/users/batch
	// GET /api/users
	// GET /api/users/:id
	// HEAD /api/users
	// HEAD /api/users/:id
	// OPTIONS /api/users
	// OPTIONS /api/users/:id
	// PATCH /api/users/:id
	// POST /api/users
	// POST /api/users/batch
	// PUT /api/users/:id
}
