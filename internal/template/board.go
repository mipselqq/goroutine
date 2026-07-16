package template

import "fmt"

type BoardCreate struct {
	Name string
}

func (t BoardCreate) String() string {
	return fmt.Sprintf("You created board '%s'", t.Name)
}

type BoardRename struct {
	Source string
	Target string
}

func (t BoardRename) String() string {
	return fmt.Sprintf("You renamed board from '%s' to '%s'", t.Source, t.Target)
}

type BoardUpdate struct {
	Name string
}

func (t BoardUpdate) String() string {
	return fmt.Sprintf("You updated board '%s'", t.Name)
}

type BoardDelete struct {
	Name string
}

func (t BoardDelete) String() string {
	return fmt.Sprintf("You deleted board '%s'", t.Name)
}
