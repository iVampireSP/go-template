package schedule

import (
	"fmt"
	"strings"
	"time"
)

// Cron 设置自定义 cron 表达式
// 格式支持 6 字段（秒 分 时 日 月 周）或 5 字段（分 时 日 月 周）
func (e *Event) Cron(expression string) *Event {
	e.expression = expression
	return e
}

// EveryMinute 每分钟执行
func (e *Event) EveryMinute() *Event {
	e.expression = "0 * * * * *"
	return e
}

// EveryTwoMinutes 每两分钟执行
func (e *Event) EveryTwoMinutes() *Event {
	e.expression = "0 */2 * * * *"
	return e
}

// EveryThreeMinutes 每三分钟执行
func (e *Event) EveryThreeMinutes() *Event {
	e.expression = "0 */3 * * * *"
	return e
}

// EveryFiveMinutes 每五分钟执行
func (e *Event) EveryFiveMinutes() *Event {
	e.expression = "0 */5 * * * *"
	return e
}

// EveryTenMinutes 每十分钟执行
func (e *Event) EveryTenMinutes() *Event {
	e.expression = "0 */10 * * * *"
	return e
}

// EveryFifteenMinutes 每十五分钟执行
func (e *Event) EveryFifteenMinutes() *Event {
	e.expression = "0 */15 * * * *"
	return e
}

// EveryThirtyMinutes 每三十分钟执行
func (e *Event) EveryThirtyMinutes() *Event {
	e.expression = "0 */30 * * * *"
	return e
}

// Hourly 每小时整点执行
func (e *Event) Hourly() *Event {
	e.expression = "0 0 * * * *"
	return e
}

// HourlyAt 每小时指定分钟执行
func (e *Event) HourlyAt(minute int) *Event {
	e.expression = fmt.Sprintf("0 %d * * * *", minute)
	return e
}

// EveryOddHour 每隔一小时执行（奇数小时）
func (e *Event) EveryOddHour() *Event {
	e.expression = "0 0 1-23/2 * * *"
	return e
}

// EveryTwoHours 每两小时执行
func (e *Event) EveryTwoHours() *Event {
	e.expression = "0 0 */2 * * *"
	return e
}

// EveryThreeHours 每三小时执行
func (e *Event) EveryThreeHours() *Event {
	e.expression = "0 0 */3 * * *"
	return e
}

// EveryFourHours 每四小时执行
func (e *Event) EveryFourHours() *Event {
	e.expression = "0 0 */4 * * *"
	return e
}

// EverySixHours 每六小时执行
func (e *Event) EverySixHours() *Event {
	e.expression = "0 0 */6 * * *"
	return e
}

// Daily 每天午夜执行
func (e *Event) Daily() *Event {
	e.expression = "0 0 0 * * *"
	return e
}

// DailyAt 每天指定时间执行
func (e *Event) DailyAt(timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	e.expression = fmt.Sprintf("0 %d %d * * *", minute, hour)
	return e
}

// TwiceDaily 每天两次执行
func (e *Event) TwiceDaily(firstHour, secondHour int) *Event {
	e.expression = fmt.Sprintf("0 0 %d,%d * * *", firstHour, secondHour)
	return e
}

// TwiceDailyAt 每天两次在指定分钟执行
func (e *Event) TwiceDailyAt(firstHour, secondHour, minute int) *Event {
	e.expression = fmt.Sprintf("0 %d %d,%d * * *", minute, firstHour, secondHour)
	return e
}

// Weekly 每周日午夜执行
func (e *Event) Weekly() *Event {
	e.expression = "0 0 0 * * 0"
	return e
}

// WeeklyOn 每周指定日期和时间执行
func (e *Event) WeeklyOn(day int, timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	e.expression = fmt.Sprintf("0 %d %d * * %d", minute, hour, day)
	return e
}

// Monthly 每月 1 日午夜执行
func (e *Event) Monthly() *Event {
	e.expression = "0 0 0 1 * *"
	return e
}

// MonthlyOn 每月指定日期和时间执行
func (e *Event) MonthlyOn(day int, timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	e.expression = fmt.Sprintf("0 %d %d %d * *", minute, hour, day)
	return e
}

// TwiceMonthly 每月两次执行
func (e *Event) TwiceMonthly(firstDay, secondDay int, timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	e.expression = fmt.Sprintf("0 %d %d %d,%d * *", minute, hour, firstDay, secondDay)
	return e
}

// LastDayOfMonth 每月最后一天执行
func (e *Event) LastDayOfMonth(timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	e.expression = fmt.Sprintf("0 %d %d 28-31 * *", minute, hour)
	e.When(isLastDayOfMonth)
	return e
}

// Quarterly 每季度第一天执行
func (e *Event) Quarterly() *Event {
	e.expression = "0 0 0 1 1,4,7,10 *"
	return e
}

// Yearly 每年 1 月 1 日执行
func (e *Event) Yearly() *Event {
	e.expression = "0 0 0 1 1 *"
	return e
}

// YearlyOn 每年指定月日和时间执行
func (e *Event) YearlyOn(month, day int, timeStr string) *Event {
	hour, minute := parseTime(timeStr)
	e.expression = fmt.Sprintf("0 %d %d %d %d *", minute, hour, day, month)
	return e
}

// Weekdays 限制在工作日执行（周一至周五）
func (e *Event) Weekdays() *Event {
	e.When(isWeekday)
	return e
}

// Weekends 限制在周末执行（周六和周日）
func (e *Event) Weekends() *Event {
	e.When(isWeekend)
	return e
}

// Sundays 限制在周日执行
func (e *Event) Sundays() *Event {
	e.When(func() bool { return isDayOfWeek(0) })
	return e
}

// Mondays 限制在周一执行
func (e *Event) Mondays() *Event {
	e.When(func() bool { return isDayOfWeek(1) })
	return e
}

// Tuesdays 限制在周二执行
func (e *Event) Tuesdays() *Event {
	e.When(func() bool { return isDayOfWeek(2) })
	return e
}

// Wednesdays 限制在周三执行
func (e *Event) Wednesdays() *Event {
	e.When(func() bool { return isDayOfWeek(3) })
	return e
}

// Thursdays 限制在周四执行
func (e *Event) Thursdays() *Event {
	e.When(func() bool { return isDayOfWeek(4) })
	return e
}

// Fridays 限制在周五执行
func (e *Event) Fridays() *Event {
	e.When(func() bool { return isDayOfWeek(5) })
	return e
}

// Saturdays 限制在周六执行
func (e *Event) Saturdays() *Event {
	e.When(func() bool { return isDayOfWeek(6) })
	return e
}

func parseTime(timeStr string) (hour, minute int) {
	parts := strings.Split(timeStr, ":")
	if len(parts) >= 1 {
		fmt.Sscanf(parts[0], "%d", &hour)
	}
	if len(parts) >= 2 {
		fmt.Sscanf(parts[1], "%d", &minute)
	}
	return
}

func isWeekday() bool {
	day := time.Now().Weekday()
	return day >= time.Monday && day <= time.Friday
}

func isWeekend() bool {
	day := time.Now().Weekday()
	return day == time.Saturday || day == time.Sunday
}

func isDayOfWeek(day int) bool {
	return int(time.Now().Weekday()) == day
}

func isLastDayOfMonth() bool {
	now := time.Now()
	tomorrow := now.AddDate(0, 0, 1)
	return tomorrow.Day() == 1
}
