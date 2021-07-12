package internal

var CodeText map[int]string

const (
	CodeSuccess      = 0
	CodeLackParam    = 8000 + iota // 缺少参数
	CodeInvalidParam               // 非法参数
	CodeAccessToken                // 获取access token 出错
	CodeVerifyToken                // 验证access token 出错
	CodeIllegalToken               // 非法token
	CodePlayerInfo                 // 获取玩家失败
)

func init() {
	CodeText = make(map[int]string)
	CodeText[CodeSuccess] = "success"
	CodeText[CodeLackParam] = "lack of param"
	CodeText[CodeInvalidParam] = "invalid param"
	CodeText[CodeVerifyToken] = "something wrong when verify token"
	CodeText[CodeIllegalToken] = "illegal token"
	CodeText[CodePlayerInfo] = "failed to get player info"
}
