package loggers

// Logger knows how to handle all the emailing events.
type Logger interface {
	Open(string) error
	Close() error
	SetTotal(uint)
	LogSkip(string, ...interface{})
	LogInfo(string, ...interface{})
	LogSent(string, ...interface{})
	LogWarn(string, ...interface{})
	LogFail(string, ...interface{})
	LogTest(string, ...interface{})
}

// o.Distributor.Progress(func(err error, email, message string) {
// 	switch err.(type) {
// 	case *distributors.ErrSkip:
// 		skipped++
// 		fmt.Print(`↷`)
// 	case nil:
// 		sent++
// 		fmt.Print(`⬟`)
// 	default:
// 		failed++
// 		fmt.Print(`⯐`)
// 		errors = append(errors, fmt.Sprintf(`FAIL %s %s.`, email, err.Error()))
// 	}
// })
