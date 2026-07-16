package template

import (
	"fmt"

	"goroutine/internal/domain"
)

type BoardCreateNotif struct {
	Name domain.BoardName
}

func (t BoardCreateNotif) String() string {
	return fmt.Sprintf("You created board '%s'", t.Name)
}

type BoardRenameNotif struct {
	Source domain.BoardName
	Target domain.BoardName
}

func (t BoardRenameNotif) String() string {
	return fmt.Sprintf("You renamed board from '%s' to '%s'", t.Source, t.Target)
}

type BoardDescriptionUpdateNotif struct {
	Name domain.BoardName
}

func (t BoardDescriptionUpdateNotif) String() string {
	return fmt.Sprintf("You updated description of board '%s'", t.Name)
}

type BoardDeleteNotif struct {
	Name domain.BoardName
}

func (t BoardDeleteNotif) String() string {
	return fmt.Sprintf("You deleted board '%s'", t.Name)
}
