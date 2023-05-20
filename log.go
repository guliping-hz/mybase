package mybase

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/sirupsen/logrus"
	"runtime"
	"strings"
)

// Colors
const (
	Reset       = "\033[0m"
	Red         = "\033[31m"
	Green       = "\033[32m"
	Yellow      = "\033[33m"
	Blue        = "\033[34m"
	Magenta     = "\033[35m"
	Cyan        = "\033[36m"
	White       = "\033[37m"
	BlueBold    = "\033[34;1m"
	MagentaBold = "\033[35;1m"
	RedBold     = "\033[31;1m"
	YellowBold  = "\033[33;1m"

	Custom     = logrus.Level(1000)
	ChangeFile = 1001
)

const TimeFmtLog = "2006/01/02 15:04:05.000" //毫秒保留3位有效数字
const TimeFmtLog2 = "15:04:05.000"           //毫秒保留3位有效数字

type logLevelMsg struct {
	level logrus.Level
	msg   string
}

var (
	logMy                    = logrus.New()
	logDir                   = "./log"
	logName                  = "log"
	logOneFileLimit          = 100 << 20 //默认 100M
	logFile         *os.File = nil
	logSaveDay               = 365 * 24 * time.Hour //默认日志保存365天
	isProduct                = true

	chanLog chan *logLevelMsg
	ctxLog  context.Context
)

func AddHook(hook logrus.Hook) {
	logMy.AddHook(hook)
}

type PrintHook struct {
}

// 过滤等级
func (imp *PrintHook) Levels() []logrus.Level {
	return []logrus.Level{logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel,
		logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel, logrus.TraceLevel, Custom}
}

// 打印
func (imp *PrintHook) Fire(entry *logrus.Entry) error {
	//fmt.Println(entry.Time.Date())
	//go time Format 必须使用这个时间 2006-01-02 15:04:05.000
	if entry.Logger.Out == os.Stderr { //如果输出还没有到其他地方，那么我们就不用再打印一遍
		return nil
	}

	if strings.HasSuffix(entry.Message, "\n") {
		entry.Message = entry.Message[0 : len(entry.Message)-1]
	}

	color := Reset
	switch entry.Level {
	case logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel:
		color = Red
	case logrus.WarnLevel, Custom:
		color = Yellow
	case logrus.InfoLevel:
		color = Green
	case logrus.DebugLevel:
		color = ""
	case logrus.TraceLevel:
		color = Magenta
	}

	if entry.Level == Custom {
		fmt.Printf("%s\n", entry.Message)
	} else {
		if len(entry.Data) != 0 {
			fmt.Printf("%s[%s]%s %s,%s%s\n", color, entry.Level, entry.Time.Format(TimeFmtLog2), entry.Message, entry.Data, Reset)
		} else {
			fmt.Printf("%s[%s]%s %s%s\n", color, entry.Level, entry.Time.Format(TimeFmtLog2), entry.Message, Reset)
		}
	}

	return nil
}

/*
*
@duration 日志时间保留最近多少天的。
*/
func initLogFile() error {
	now := time.Now()
	logFilePath := fmt.Sprintf("%s/%s-%s.log", logDir, logName, now.Format(DateFmtDB))
	remFilePath := fmt.Sprintf("%s/%s-%s.log", logDir, logName, now.Add(-logSaveDay).Format(DateFmtDB))
	_ = os.Remove(remFilePath)

	// You could set this to any `io.Writer` such as a file
	//D("logFile =", logFile)
	if logFile != nil { //检查之前的File
		//D("True")
		_ = logFile.Close()
	} else {
		//D("False")
	}

	logFileHere, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logMy.Errorf("创建/打开日志文件[%s]失败: err=%s\n", logFilePath, err)
	} else {
		logMy.Out = logFileHere
	}
	return err
}

/*
*
@day 保留多少天的日志
*/
func initLogDir(dir, fileName string, day int, ctx context.Context) error {
	//var osname = string(runtime.GOOS)
	//fmt.Println("os is", osname)
	//dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	//if err != nil {
	//	return err
	//}
	//if runtime.GOOS == "linux" && logDirCfg != "" && isInProduct { //linux需要制定一个比较大的磁盘
	//	dir = logDirCfg + "/" + dir
	//}
	logDir = dir + "/log"
	logName = fileName
	logSaveDay = time.Duration(day) * 24 * time.Hour

	if err := os.MkdirAll(logDir, os.ModeDir); err != nil {
		return err
	}
	ctxLog = ctx
	if chanLog == nil {
		chanLog = make(chan *logLevelMsg, 10)
		go func() {
			defer close(chanLog)
			for {
				select {
				case l := <-chanLog:
					switch l.level {
					// PanicLevel level, highest level of severity. Logs and then calls panic with the
					// message passed to Debug, Info, ...
					case logrus.PanicLevel:
						// FatalLevel level. Logs and then calls `logger.Exit(1)`. It will exit even if the
						// logging level is set to Panic.
						logMy.Panicln(l.msg)
					case logrus.FatalLevel:
						// ErrorLevel level. Logs. Used for errors that should definitely be noted.
						// Commonly used for hooks to send errors to an error tracking service.
						logMy.Fatalln(l.msg)
					case logrus.ErrorLevel:
						// WarnLevel level. Non-critical entries that deserve eyes.
						logMy.Errorln(l.msg)
					case logrus.WarnLevel:
						// InfoLevel level. General operational entries about what's going on inside the
						// application.
						logMy.Warnln(l.msg)
					case logrus.InfoLevel:
						// DebugLevel level. Usually only enabled when debugging. Very verbose logging.
						logMy.Infoln(l.msg)
					case logrus.DebugLevel:
						// TraceLevel level. Designates finer-grained informational events than the Debug.
						logMy.Debugln(l.msg)
					case logrus.TraceLevel:
						logMy.Traceln(l.msg)
					case Custom: //自定义统一为warning
						logMy.Warnln(l.msg)
					case ChangeFile:
						_ = initLogFile()
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	return initLogFile()
}

func InitLogModule(dir, fileName string, day int, isInProduct bool, debugLv logrus.Level, ctx context.Context) error {
	// do something here to set environment depending on an environment variable
	// or command-line flag
	isProduct = isInProduct
	if isInProduct {
		//logMy.Formatter = &logrus.JSONFormatter{
		//	TimestampFormat: TimeFmtLog,
		//} //为了方便使用Logstash
		logMy.Formatter = &logrus.TextFormatter{
			TimestampFormat: TimeFmtLog2,
		}
		logMy.SetLevel(logrus.InfoLevel)
	} else {
		logMy.Formatter = &logrus.TextFormatter{
			TimestampFormat: TimeFmtLog2,
		}
		//logMy.SetLevel(logrus.TraceLevel)
		logMy.SetLevel(debugLv)
	}

	InitNoFile()
	return initLogDir(dir, fileName, day, ctx)
}

func InitNoFile() {
	//添加监听Hook
	logMy.Hooks.Add(new(PrintHook))
}

func CheckDay() {
	select {
	case <-ctxLog.Done():
	default:
		chanLog <- &logLevelMsg{level: ChangeFile}
	}
}

func wrapFormat(format, file string, line int) string {
	formatNew := fmt.Sprintf("%s:%d %s", file, line, format)
	return formatNew
}

func toLog(lv logrus.Level, msg string) {
	if ctxLog == nil {
		return
	}

	select {
	case <-ctxLog.Done():
	default:
		chanLog <- &logLevelMsg{level: lv, msg: msg}
	}
}

func D(format string, args ...interface{}) {
	//0的话获取的是129行调用，我们要获取外层调用的位置
	if isProduct {
		return
	}

	_, file, line, ok := runtime.Caller(1)
	if ok {
		toLog(logrus.DebugLevel, fmt.Sprintf(wrapFormat(format, file, line), args...))
	}
}

func I(format string, args ...interface{}) {
	//0的话获取的是129行调用，我们要获取外层调用的位置
	_, file, line, ok := runtime.Caller(1)
	if ok {
		toLog(logrus.InfoLevel, fmt.Sprintf(wrapFormat(format, file, line), args...))
	}
}

func W(format string, args ...interface{}) {
	//0的话获取的是129行调用，我们要获取外层调用的位置
	_, file, line, ok := runtime.Caller(1)
	if ok {
		toLog(logrus.WarnLevel, fmt.Sprintf(wrapFormat(format, file, line), args...))
	}
}

func E(format string, args ...interface{}) {
	//0的话获取的是129行调用，我们要获取外层调用的位置
	_, file, line, ok := runtime.Caller(1) //funcName
	//fmt.Println("Func Name=" + runtime.FuncForPC(funcName).Name())
	if ok {
		toLog(logrus.ErrorLevel, fmt.Sprintf(wrapFormat(format, file, line), args...))
	}
}

func P(format string, args ...interface{}) {
	if isProduct {
		return
	}

	//0的话获取的是129行调用，我们要获取外层调用的位置
	_, file, line, ok := runtime.Caller(1) //funcName
	//fmt.Println("Func Name=" + runtime.FuncForPC(funcName).Name())
	if ok {
		toLog(logrus.PanicLevel, fmt.Sprintf(wrapFormat(format, file, line), args...))
	}
}

// 用于追踪定位
func T(format string, args ...interface{}) {
	stack := debug.Stack()
	_, file, line, ok := runtime.Caller(1)
	if ok {
		toLog(logrus.WarnLevel, fmt.Sprintf(wrapFormat(format, file, line), args...)+"\n\n\nstack="+string(stack))
	}
}

// 自定义打印，兼容其他依赖
func C(p []byte) {
	toLog(Custom, string(p))
}

type LogWriter struct {
}

func (w *LogWriter) Write(p []byte) (n int, err error) {
	C(p)
	return len(p), nil
}
