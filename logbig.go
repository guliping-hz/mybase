package mybase

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"
	"sync"
	"time"
)

var (
	logBigMutex    sync.Mutex
	logCurFileName string
	logBigFile     *os.File

	logBigDir       string
	logBigPrefix    string
	logBigLimitSize int64         = 10 << 20
	logBigLimitDay  time.Duration = 3
)

func InitLogBigFile(isProd bool, dir, prefix string, limitSize int64, limitDay time.Duration) error {
	//if runtime.GOOS == "linux" && isProd { //linux需要制定一个比较大的磁盘
	//	linuxDetail, err := filepath.Abs(filepath.Dir(os.Args[0]))
	//	if err != nil {
	//		return err
	//	}
	//	dir = linuxDir + linuxDetail //更换目录
	//}

	logBigDir = dir + "/logbig"
	logBigPrefix = prefix
	logBigLimitSize = limitSize
	logBigLimitDay = limitDay

	if err := os.MkdirAll(logBigDir, 0744); err != nil {
		return err
	}

	return initLogBigFile()
}

func initLogBigFile() error {
	entrys, err := os.ReadDir(logBigDir)
	if err != nil {
		return err
	}

	now := time.Now()
	limitDayFileName := fmt.Sprintf("%s-%s-", logBigPrefix, now.Add(-time.Hour*24*logBigLimitDay).Format(DateFmtDB))
	filePattern := fmt.Sprintf("%s-%s-", logBigPrefix, now.Format(DateFmtDB))
	//fmt.Println(limitDayFileName, filePattern)
	expToday, _ := regexp.Compile(filePattern + `(\d+).log$`)
	curFileName := ""
	seq := int64(0)
	var curFileInfo os.FileInfo
	//从当前的日志列表中找出最大的文件seq
	for i := range entrys {
		curFileName = entrys[i].Name()
		if curFileName < limitDayFileName {
			_ = os.Remove(path.Join(logBigDir, curFileName))
			continue
		}

		curFileInfo, _ = entrys[i].Info()
		//fmt.Println(i, entrys[i].Name(), curFileInfo.ModTime().Format(TimeFmtDB))
		finds := expToday.FindStringSubmatch(curFileName)
		if finds != nil {
			seq, _ = strconv.ParseInt(finds[1], 10, 32)
		}
	}

	if curFileInfo != nil && curFileInfo.Size() > logBigLimitSize {
		seq++
	}
	curFileName = filePattern + fmt.Sprintf("%03d.log", seq)
	//fmt.Println(curFileName)
	if logCurFileName != curFileName {
		logCurFileName = curFileName
		logBigMutex.Lock()
		defer logBigMutex.Unlock()
		if logBigFile != nil {
			_ = logBigFile.Close()
		}
		logBigFile, _ = os.OpenFile(logBigDir+"/"+logCurFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	}
	return nil
}

func CheckDayForLogBig() {
	_ = initLogBigFile()
}

// 内部不提供换行，需要自己换行。
func LogBig(format string, args ...any) {
	logBigMutex.Lock()
	defer logBigMutex.Unlock()
	if logBigFile == nil {
		return
	}

	if info, err := logBigFile.Stat(); err == nil {
		if info.Size() >= logBigLimitSize {
			logBigFile.Close()
			logBigMutex.Unlock() //解锁一下
			initLogBigFile()
			logBigMutex.Lock()
		}
	}
	_, _ = logBigFile.WriteString(time.Now().Format(TimeFmtLog2) + " ") //时间
	_, _ = fmt.Fprintf(logBigFile, format, args...)                     //写入内容
}
