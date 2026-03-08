package usagerule

import (
	"fmt"
	"time"
)

// TimeGranularity 表示时间窗口的粒度/周期类型。
type TimeGranularity string

// String returns the string representation of the time granularity.
func (r TimeGranularity) String() string {
	return string(r)
}

// IsValid checks if the time granularity is one of the defined constants.
func (r TimeGranularity) IsValid() bool {
	switch r {
	case GranularityMinute, GranularityHour,
		GranularityDay, GranularityWeek,
		GranularityMonth:
		return true
	default:
		return false
	}
}

const (
	GranularityMinute TimeGranularity = "minute"
	GranularityHour   TimeGranularity = "hour"
	GranularityDay    TimeGranularity = "day"
	GranularityWeek   TimeGranularity = "week"
	GranularityMonth  TimeGranularity = "month"
)

type SourceType int

const (
	// SourceTypeToken token数量/积分数量
	SourceTypeToken SourceType = 1
	// SourceTypeRequest 请求次数
	SourceTypeRequest SourceType = 2
)

// UsageRule 用量规则定义
type UsageRule struct {
	SourceType      SourceType      `json:"source_type"`
	TimeGranularity TimeGranularity `json:"time_granularity"`
	// WindowSize 窗口大小,单位为时间粒度。
	// 比如时间粒度为分钟，窗口大小为10，则窗口为10分钟。
	WindowSize int `json:"window_size"`
	// Total 总量
	Total float64 `json:"total"`
}

// UsageStats 规则的用量统计信息
type UsageStats struct {
	// Rule 关联的用量规则
	Rule *UsageRule `json:"rule"`
	// Used 已使用量
	Used float64 `json:"used"`
	// Remain 剩余量
	Remain float64 `json:"remain"`
	// StartTime 窗口开始时间
	StartTime *time.Time `json:"start_time"`
	// EndTime 窗口结束时间
	EndTime *time.Time `json:"end_time"`
}

// IsValid 判断规则是否有效（非 nil 且 Total > 0）
func (r *UsageRule) IsValid() bool {
	return r != nil && r.Total > 0
}

// String 返回规则字符串
func (r *UsageRule) String() string {
	return fmt.Sprintf("source_type: %d, time_granularity: %s, window_size: %d, total: %f",
		r.SourceType, r.TimeGranularity, r.WindowSize, r.Total)
}

// Clone 克隆规则
func (r *UsageRule) Clone() *UsageRule {
	return &UsageRule{
		SourceType:      r.SourceType,
		TimeGranularity: r.TimeGranularity,
		WindowSize:      r.WindowSize,
		Total:           r.Total,
	}
}

// CalculateWindowTime 根据当前时间，计算窗口开始时间和结束时间
func (r *UsageRule) CalculateWindowTime() (start, end *time.Time) {
	now := time.Now()
	var s time.Time

	switch r.TimeGranularity {
	case GranularityMinute:
		// 对齐到分钟起始
		s = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())
		e := s.Add(time.Duration(r.WindowSize) * time.Minute)
		return &s, &e
	case GranularityHour:
		// 对齐到小时起始
		s = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, now.Location())
		e := s.Add(time.Duration(r.WindowSize) * time.Hour)
		return &s, &e
	case GranularityDay:
		// 对齐到天起始
		s = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		e := s.AddDate(0, 0, r.WindowSize)
		return &s, &e
	case GranularityWeek:
		// 对齐到本周一起始
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7 // 将周日从0改为7，使周一为1
		}
		s = time.Date(now.Year(), now.Month(), now.Day()-weekday+1, 0, 0, 0, 0, now.Location())
		e := s.AddDate(0, 0, r.WindowSize*7)
		return &s, &e
	case GranularityMonth:
		// 对齐到月起始
		s = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		e := s.AddDate(0, r.WindowSize, 0)
		return &s, &e
	}
	return nil, nil
}

// IsTriggered 检查是否触发了限流
func (s *UsageStats) IsTriggered() bool {
	return s.Remain <= 0
}

// IsInWindow 判断当前时间是否在窗口内
func (s *UsageStats) IsInWindow() bool {
	t := time.Now()
	if s.StartTime == nil || s.EndTime == nil {
		return false
	}
	return t.After(*s.StartTime) && t.Before(*s.EndTime)
}

// String 返回用量统计字符串
func (s *UsageStats) String() string {
	ruleStr := ""
	if s.Rule != nil {
		ruleStr = s.Rule.String()
	}
	return fmt.Sprintf("%s, used: %f, remain: %f, start_time: %v, end_time: %v",
		ruleStr, s.Used, s.Remain, s.StartTime, s.EndTime)
}

// Clone 克隆用量统计
func (s *UsageStats) Clone() *UsageStats {
	clone := &UsageStats{
		Used:      s.Used,
		Remain:    s.Remain,
		StartTime: s.StartTime,
		EndTime:   s.EndTime,
	}
	if s.Rule != nil {
		clone.Rule = s.Rule.Clone()
	}
	return clone
}

// CalculateNextWindowTime 计算下一个周期的开始时间和结束时间
func (s *UsageStats) CalculateNextWindowTime() (*time.Time, *time.Time) {
	if s.Rule == nil {
		return nil, nil
	}
	// 若当前窗口尚未计算，先计算当前窗口
	if s.EndTime == nil {
		start, end := s.Rule.CalculateWindowTime()
		s.StartTime = start
		s.EndTime = end
	}
	if s.EndTime == nil {
		return nil, nil
	}

	// 以当前窗口结束时间作为下一个窗口的起始点
	nextStart := *s.EndTime
	var nextEnd time.Time

	switch s.Rule.TimeGranularity {
	case GranularityMinute:
		nextEnd = nextStart.Add(time.Duration(s.Rule.WindowSize) * time.Minute)
	case GranularityHour:
		nextEnd = nextStart.Add(time.Duration(s.Rule.WindowSize) * time.Hour)
	case GranularityDay:
		nextEnd = nextStart.AddDate(0, 0, s.Rule.WindowSize)
	case GranularityWeek:
		nextEnd = nextStart.AddDate(0, 0, s.Rule.WindowSize*7)
	case GranularityMonth:
		nextEnd = nextStart.AddDate(0, s.Rule.WindowSize, 0)
	default:
		return nil, nil
	}

	return &nextStart, &nextEnd
}
