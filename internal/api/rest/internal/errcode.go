package internal

var CodeText = make(map[int]string)

// Code 定义规则：每个服务的起始值为 serverId * 10000 + 业务ID
const (
	CodeSuccess      = 0
	CodeLackParam    = 8000 + iota // 缺少参数
	CodeInvalidParam               // 非法参数
	CodeAccessToken                // 获取 token 出错
	CodeVerifyToken                // 验证 token 出错
	CodeIllegalToken               // 非法 token
	CodePlayerInfo                 // 获取玩家失败
)

func init() {
	CodeText[CodeSuccess] = "ok"
	CodeText[CodeLackParam] = "lack of param"
	CodeText[CodeInvalidParam] = "invalid param"
	CodeText[CodeVerifyToken] = "something wrong when verify token"
	CodeText[CodeIllegalToken] = "illegal token"
	CodeText[CodePlayerInfo] = "failed to get player info"
}
