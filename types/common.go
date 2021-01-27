package types

type Subscription struct {
	Name   string
	URL    string
	Rule   string
	Enable bool
}

type CustomOutbound struct {
	Ps       string
	Filename string
}

type ListMode int

func (l ListMode) String() string {
	switch l {
	case ModeNone:
		return "ModeNone"
	case ModeBlackList:
		return "ModeBlackList"
	case ModeWhiteList:
		return "ModeWhiteList"
	}
	return "Unknown mode"
}

const (
	ModeNone ListMode = iota
	ModeWhiteList
	ModeBlackList
)

type FilterConfig struct {
	Mode ListMode
	// 黑名单或者白名单匹配时的keyword
	Lists []string
}
