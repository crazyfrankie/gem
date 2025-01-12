package binding

import "net/http"

type plainBinding struct {
}

func (p plainBinding) Name() string {
	//TODO implement me
	panic("implement me")
}

func (p plainBinding) Bind(request *http.Request, a any) error {
	//TODO implement me
	panic("implement me")
}

func (p plainBinding) BindBody(bytes []byte, a any) error {
	//TODO implement me
	panic("implement me")
}
