package app

import (
	"encoding/json"
	"time"
)

type Order struct {
	Timestamp time.Time `json:"timestamp"`
	BlobId    string    `json:"blobId"`
}

func (o Order) Marshal() ([]byte, error) {
	return json.Marshal(&o)
}

func UnmarshalOrder(raw []byte) (Order, error) {
	o := Order{}

	err := json.Unmarshal(raw, &o)

	return o, err
}
