package internal

// Response ...
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type HellReq struct {
	Name string `json:"name"`
}

type HelloRsp struct {
	Greet string `json:"greet"`
}
