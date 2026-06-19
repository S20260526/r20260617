package domain

type Hello struct {
}

type World struct {
}

func NewHello() Hello {
	return Hello{}
}

func NewWorld() World {
	return World{}
}

func (h Hello) String() string {
	return "Hello"
}

func (w World) String() string {
	return "World"
}
