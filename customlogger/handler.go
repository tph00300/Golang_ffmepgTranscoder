package customlogger

import (
	"log"
	"os"
)

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	CritLogger    *log.Logger
	DebugLogger   *log.Logger
)

func Init() { // 로그파일을 열고, 각 로그파일을 관리하는 변수들을 초기화 시켜준다.
	// 로그 파일을 0666 권한으로 연다.
	file, err := os.OpenFile("/var/log/transcode.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

	// 로그파일을 못 열을 경우
	if err != nil {
		log.Fatal(err)
	}
	// InfoLogger를 통해서 출력하게 되면 INFO: 2016/01/15 15:30:53 test.go:27: MsgContent 이런식으로 출력
	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	// WarningLogger를 통해서 출력하게 되면 WARNING: 2016/01/15 15:30:53 test.go:27: MsgContent 이런식으로 출력
	WarningLogger = log.New(file, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	// CritLogger를 통해서 출력하게 되면 CRIT: 2016/01/15 15:30:53 test.go:27: MsgContent 이런식으로 출력
	CritLogger = log.New(file, "CRIT: ", log.Ldate|log.Ltime|log.Lshortfile)

	DebugLogger = log.New(file, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

}
