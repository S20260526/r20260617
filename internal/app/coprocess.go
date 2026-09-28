package app

import (
	c "context"
	"internal/grpcipc"
)

type CoprocessRequest = grpcipc.Request
type CoprocessResponse = grpcipc.Response
type CoprocessResult = grpcipc.Result

const CoprocessResultNo = grpcipc.Result_RESULT_NO
const CoprocessResultYes = grpcipc.Result_RESULT_YES
const CoprocessResultFail = grpcipc.Result_RESULT_FAIL

type Ipc interface {
	SocketName() string
	Call(ctx c.Context, request *CoprocessRequest) (*CoprocessResponse, error)
}

type Coprocess interface {
	Call(ctx c.Context, request *CoprocessRequest) (*CoprocessResponse, error)
	Wait() error
}
