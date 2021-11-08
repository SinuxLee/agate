package internal

// Response ...
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type LoginReq struct {
	UserName string `json:"userName"`
	Password string `json:"password"`
}

type LoginRsp struct {
	Token    string `json:"token"`
	ExpireIn int    `json:"expireIn"`
}

type HellReq struct {
	Name string `json:"name"`
}

type HelloRsp struct {
	Greet string `json:"greet"`
}
