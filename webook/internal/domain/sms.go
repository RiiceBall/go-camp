package domain

type Sms struct {
	Id        int64
	TplId     string
	Args      []string
	Numbers   []string
	RetryLeft int // 剩余重试次数
}
