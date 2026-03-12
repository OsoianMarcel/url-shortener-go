package domain

import "context"

// HealthDependency is a port for any dependency whose availability can be checked.
type HealthDependency interface {
	Name() string
	Ping(ctx context.Context) error
}
