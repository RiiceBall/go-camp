package domain

// WechatInfo 微信的授权信息
type WechatInfo struct {
	// OpenId 是应用内唯一
	OpenId string
	// UnionId 是整个公司账号内唯一
	UnionId string
}
