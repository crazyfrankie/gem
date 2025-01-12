package binding

import "net/http"

type queryBinding struct {
}

func (q queryBinding) Name() string {
	//TODO implement me
	panic("implement me")
}

func (q queryBinding) Bind(request *http.Request, a any) error {
	//TODO implement me
	panic("implement me")
}
