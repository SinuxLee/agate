package internal

// Response ...
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type HellReq struct {
	Name string `json:"name"`
}

type HelloRsp struct {
	Greet string `json:"greet"`
}
