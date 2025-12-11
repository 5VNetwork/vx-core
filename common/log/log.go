package log

// var (
// 	WarningLogger *log.Logger
// 	InfoLogger    *log.Logger
// 	ErrorLogger   *log.Logger
// 	DebugLogger   *log.Logger
// 	SevereLogger  *log.Logger

// 	Level LogLevel
// )

// func getFilepath(depth int) string {
// 	_, file, line, ok := runtime.Caller(depth)
// 	if !ok {
// 		file = "???"
// 		line = 0
// 	}
// 	short := file
// 	time := 0
// 	for i := len(file) - 1; i > 0; i-- {
// 		if file[i] == '/' {
// 			time++
// 			if time == 2 {
// 				short = file[i+1:]
// 				break
// 			}
// 		}
// 	}
// 	file = short
// 	return file + ":" + fmt.Sprint(line)
// }

// func Printf(format string, v ...any) {
// 	log.Printf(format, v...)
// }

// func Println(v ...any) {
// 	log.Println(v...)
// }

// func PrintDebug(values ...interface{}) {
// 	if Level <= LogLevel_DEBUG {
// 		DebugLogger.Println(buildString(values...))
// 	}
// }

// func PrintfDebug(format string, v ...any) {
// 	if Level <= LogLevel_DEBUG {
// 		s := fmt.Sprintf(format, v...)
// 		DebugLogger.Println(buildString(s))
// 	}
// }

// func PrintError(values ...interface{}) {
// 	if Level <= LogLevel_ERROR {
// 		ErrorLogger.Println(buildString(values...))
// 	}
// }

// func PrintfError(format string, v ...any) {
// 	if Level <= LogLevel_ERROR {
// 		s := fmt.Sprintf(format, v...)
// 		ErrorLogger.Println(buildString(s))
// 	}
// }

// func PrintWarn(values ...interface{}) {
// 	if Level <= LogLevel_WARNING {
// 		WarningLogger.Println(buildString(values...))
// 	}
// }

// func PrintfWarn(format string, v ...any) {
// 	if Level <= LogLevel_WARNING {
// 		s := fmt.Sprintf(format, v...)
// 		WarningLogger.Println(buildString(s))
// 	}
// }

// func PrintInfo(values ...interface{}) {
// 	if Level <= LogLevel_INFO {
// 		InfoLogger.Println(buildString(values...))
// 	}
// }

// func buildString(values ...interface{}) string {
// 	return serial.Concat(values...) + " (" + getFilepath(3) + ")"
// }

// func PrintfSevere(format string, v ...any) {
// 	if Level <= LogLevel_SEVERE {
// 		s := fmt.Sprintf(format, v...)
// 		SevereLogger.Println(buildString(s))
// 	}
// }

// func init() {
// 	Level = LogLevel_NONE
// }
