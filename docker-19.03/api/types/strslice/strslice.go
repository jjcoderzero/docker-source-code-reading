package strslice // import "github.com/docker/docker/api/types/strslice"

import "encoding/json"

// StrSlice表示一个字符串或字符串数组。我们需要覆盖json解码器以接受这两个选项。
type StrSlice []string

// UnmarshalJSON对字节片进行解码，不管它是字符串还是字符串数组。要实现json.Unmarshaler需要此方法。
func (e *StrSlice) UnmarshalJSON(b []byte) error {
	if len(b) == 0 {
		// 在没有输入的情况下，我们通过返回nil和保留目标来保留现有的值。这允许为类型定义默认值。
		return nil
	}

	p := make([]string, 0, 1)
	if err := json.Unmarshal(b, &p); err != nil {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		p = append(p, s)
	}

	*e = p
	return nil
}
