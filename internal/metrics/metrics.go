package metrics

import (
	"expvar"
	"runtime"
	"time"
)

var (
	BuildTime string
	Version   string
)

func init() {
	expvar.NewString("version").Set(Version)

	//expvar.Publish("database", expvar.Func(func() interface{} {
	//	return db.Stats()
	//}))

	expvar.Publish("goroutines", expvar.Func(func() interface{} {
		return runtime.NumGoroutine()
	}))

	expvar.Publish("timestamp", expvar.Func(func() interface{} {
		return time.Now().Format("2006-01-02T15:04:05")
	}))
}
