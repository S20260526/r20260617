package app

import (
	"testing"
	"time"
)

func TestOrderMarshaling(t *testing.T) {
	in := Order{
		Timestamp: time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC),
		BlobId:    "1234",
	}

	raw, err := in.Marshal()

	t.Log(in.Timestamp, string(raw), err)

	if err != nil {
		t.Fatal(err)
	}

	out, err := UnmarshalOrder(raw)

	if err != nil || out != in {
		t.Error()
	}
}
