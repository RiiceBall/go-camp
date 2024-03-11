package web

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

type ErrorResult struct {
	Result   Result
	ErrorMsg string
	Err      error
	Context  map[string]interface{}
}

func (e ErrorResult) Error() string {
	return e.ErrorMsg
}
