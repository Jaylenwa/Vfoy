package requestmock

import (
	"io"

	"github.com/Jaylenwa/Vfoy/pkg/request"
	"github.com/stretchr/testify/mock"
)

type RequestMock struct {
	mock.Mock
}

func (r RequestMock) Request(method, target string, body io.Reader, opts ...request.Option) *request.Response {
	return r.Called(method, target, body, opts).Get(0).(*request.Response)
}
