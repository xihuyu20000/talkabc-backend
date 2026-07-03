package logger

import (
	"fmt"
	"time"
)

type GormLogger struct{}

func NewGormLogger() *GormLogger {
	return &GormLogger{}
}

func (l *GormLogger) Print(values ...interface{}) {
	if len(values) == 0 {
		return
	}

	switch values[0] {
	case "sql":
		if len(values) >= 6 {
			caller := fmt.Sprintf("%v", values[1])
			elapsed := values[2].(time.Duration)
			sql := fmt.Sprintf("%v", values[3])
			vars := values[4]
			rows := values[5].(int64)

			Debugf("[GORM] Caller: %s, Elapsed: %v, Rows: %d, SQL: %s, Vars: %v",
				caller, elapsed, rows, sql, vars)
		} else {
			Debugf("[GORM] SQL: %v", values)
		}
	case "log":
		if len(values) >= 2 {
			Infof("[GORM] %v", values[1])
		}
	default:
		Debugf("[GORM] %v", values)
	}
}
