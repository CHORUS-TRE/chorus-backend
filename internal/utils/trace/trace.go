package trace

import (
	"runtime"
	"strings"
)

func Caller() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return strings.TrimPrefix(frame.Function, "github.com/CHORUS-TRE/chorus-backend")
}
