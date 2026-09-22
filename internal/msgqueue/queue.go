package msgqueue

type Queue interface {
	Connect(url string) error
	OpenChannel() error
	CloseChannel()
	Disconnect()
	Publish(msg []byte) error
	Consume() ([]byte, error)
}
