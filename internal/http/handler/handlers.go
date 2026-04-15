package handler

type Handlers struct {
	Auth    *Auth
	Health  *Health
	Boards  *Boards
	Columns *Columns
}

func NewHandlers(auth *Auth, health *Health, boards *Boards, columns *Columns) *Handlers {
	return &Handlers{
		Auth:    auth,
		Health:  health,
		Boards:  boards,
		Columns: columns,
	}
}
