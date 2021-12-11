package internal

// Response ...
type Response struct {
	Code    int         `json:"code"`           // < 0表示框架层面错误码; =0 表示成功; >0 表示业务层错误码
	Message string      `json:"message"`        // code 非零时，返回错误原因
	Data    interface{} `json:"data,omitempty"` // code 为零时，返回业务数据
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
