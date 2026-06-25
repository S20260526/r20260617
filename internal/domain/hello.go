package domain

type Hello struct {
}

func NewHello() Hello {
	return Hello{}
}

func (h Hello) String() string {
	return "Hello"
}
