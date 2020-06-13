package user

type session interface {
	UpdateCurrentUserPasswordAllowed() bool
}