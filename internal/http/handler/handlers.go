package handler

type Handlers struct {
	Auth   *Auth
	Health *Health
}

func NewHandlers(auth *Auth, health *Health) *Handlers {
	return &Handlers{
		Auth:   auth,
		Health: health,
	}
}
