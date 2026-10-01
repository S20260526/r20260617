package app

import (
	"context"
	"errors"
	"internal/app/msgqueue"
	"internal/grpcipc"
	"net/http"
	"time"
)

var fail = errors.New("fail")
var tstmp = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

type mockIncoming struct {
	payload      string
	acknowledged string
}

func (mi *mockIncoming) GetData() []byte {
	return []byte(mi.payload)
}

func (mi *mockIncoming) Acknowledge() {
	mi.acknowledged = "A"
}

func (mi *mockIncoming) Reject() {
	mi.acknowledged = "R"
}

type mockQueue struct {
	fail    bool
	trace   string
	url     string
	payload string
	in      *mockIncoming
}

func (q *mockQueue) Connect(url string) error {
	q.trace += "C"
	q.url = url

	return nil
}

func (q *mockQueue) OpenChannel() error {
	q.trace += "O"

	return nil
}
func (q *mockQueue) CloseChannel() {
	q.trace += "X"
}

func (q *mockQueue) Disconnect() {
	q.trace += "D"
}

func (q *mockQueue) Publish(ctx context.Context, msg []byte) error {
	q.trace += "P"
	q.payload = string(msg)

	if q.fail {
		return fail
	}

	return nil
}

func (q *mockQueue) Consume(ctx context.Context) (msgqueue.Incoming, error) {
	q.trace += "G"

	if q.fail {
		return nil, fail
	}

	q.in = &mockIncoming{payload: q.payload}

	return q.in, nil
}

type mockStorage struct {
	fail    bool
	trace   string
	key     string
	payload string
}

func (s *mockStorage) Create(ctx context.Context, blob []byte) (string, error) {
	s.trace += "C"

	if s.fail {
		return "", fail
	}

	s.payload = string(blob)

	return s.key, nil
}

func (s *mockStorage) Read(ctx context.Context, key string) ([]byte, error) {
	s.trace += "R"

	s.key = key

	if s.fail {
		return nil, fail
	}

	return []byte(s.payload), nil
}

func (s *mockStorage) Delete(ctx context.Context, key string) error {
	s.trace += "D"

	s.key = key

	if s.fail {
		return fail
	}

	return nil
}

type mockCoprocess struct {
	trace   string
	payload string
	failed  bool
	result  CoprocessResult
}

func (c *mockCoprocess) Call(ctx context.Context, request *CoprocessRequest) (*CoprocessResponse, error) {
	c.trace += "C"
	c.payload = string(request.Payload)

	if c.failed {
		return nil, fail
	}

	return &CoprocessResponse{Result: grpcipc.Result(c.result)}, nil
}

func (c *mockCoprocess) Wait() error {
	return fail
}

type mockRegistry struct {
	trace  string
	table  string
	recent Event
	failed bool
}

func (r *mockRegistry) Put(ctx context.Context, table string, event Event) error {
	r.trace += "P"
	r.table = table
	r.recent = event

	if r.failed {
		return fail
	}

	return nil
}

func (r *mockRegistry) Cleanup() {
}

type mockMetrics struct {
	in, out, yes, no, fail, err int
	beginTime                   time.Time
}

func (m *mockMetrics) RegIn() {
	m.in++
}

func (m *mockMetrics) RegOut() {
	m.out++
}

func (m *mockMetrics) RegYes() {
	m.yes++
}
func (m *mockMetrics) RegNo() {
	m.no++
}

func (m *mockMetrics) RegFail() {
	m.fail++
}

func (m *mockMetrics) RegErr() {
	m.err++
}

func (m *mockMetrics) RegFrontToEndDuration(beginTime time.Time) {
	m.beginTime = beginTime
}

func (m *mockMetrics) GetHttpHandler() http.Handler {
	return nil
}
