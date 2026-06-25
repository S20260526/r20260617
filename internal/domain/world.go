package domain

type World struct {
}

func NewWorld() World {
	return World{}
}

func (w World) String() string {
	return "World"
}
