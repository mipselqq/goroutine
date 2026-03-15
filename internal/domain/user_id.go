package domain

type (
	userID struct{}
	UserID = UUID[userID]
)

func NewUserID() UserID {
	return NewID[userID]()
}

func ParseUserID(s string) (UserID, error) {
	return ParseID[userID](s)
}
