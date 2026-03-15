package handler

type Handlers struct {
	Auth   *Auth
	Health *Health
	Boards *Boards
}

func NewHandlers(auth *Auth, health *Health, boards *Boards) *Handlers {
	return &Handlers{
		Auth:   auth,
		Health: health,
		Boards: boards,
	}
}
