package k8log

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"time"

	"github.com/fatih/color"
)

// 日志级别对应的颜色分别是
// Info: 绿色
// Error: 红色
// Warn: 黄色
// Debug: 蓝色
// Fatal: 红色

var ifDebug = true
var ifDirectPrint = true
var logPath = ""
var globalLogFile *os.File = nil

const InfoLogFormat string = "[%s]: [%s] %s\n"
const ErrorLogFormat string = "[%s]: [%s] [func From: %s] [file: %s] [line: %d] %s\n"
const WarnLogFormat string = "[%s]: [%s] %s\n"
const DebugLogFormat string = "[%s]: [%s] [func From: %s] [file: %s] [line: %d] %s\n"
const FatalLogFormat string = "[%s]: [%s] [func From: %s] [file: %s] [line: %d] %s\n"

// 注意！这个是这个包被加载的时候要做的事情
func init() {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}

	// 如果k8s目录不存在就创建一个目录
	if _, err := os.Stat(usr.HomeDir + "/k8s"); os.IsNotExist(err) {
		err := os.Mkdir(usr.HomeDir+"/k8s", 0755)
		if err != nil {
			panic(err)
		}
	}

	logPath = usr.HomeDir + "/k8s" + "/k8s.log"
	// 初始化的时候 检查文件是否存在，不存在就创建一个
	// 处理完成后关闭文件，避免不必要的文件资源占用
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	globalLogFile = logFile
	if err != nil {
		panic(err)
	}
}

// 基本的日志信息
func InfoLog(component string, msg string) {
	// 1. 获取当前的时间
	t := time.Now()
	currentTimeStr := t.Format("2006-01-02 15:04:05")
	// 2. 根据InfoLogFormat格式化字符串
	logStr := fmt.Sprintf(InfoLogFormat, currentTimeStr, component, msg)
	// 3. 将字符串写入到文件中
	globalLogFile.WriteString(logStr)

	if ifDirectPrint {
		// 4. 将字符串打印到控制台
		color.Green(logStr)
	}
}

// 错误日志
func ErrorLog(component string, msg string) {
	// 1. 获取当前的时间
	t := time.Now()
	currentTimeStr := t.Format("2006-01-02 15:04:05")
	// 2. 获取发起调用的函数名，文件名，行号
	funcName, file, line, _ := runtime.Caller(1)

	// 3. 根据 ErrorLogFormat 组装成字符串
	logStr := fmt.Sprintf(ErrorLogFormat, currentTimeStr, component, runtime.FuncForPC(funcName).Name(), file, line, msg)
	// 4. 将字符串写入到文件中
	globalLogFile.WriteString(logStr)

	if ifDirectPrint {
		// 5. 将字符串打印到控制台
		color.Red(logStr)
	}
}

// 警告日志
func WarnLog(component string, msg string) {
	// 1. 获取当前的时间
	t := time.Now()
	currentTimeStr := t.Format("2006-01-02 15:04:05")
	// 2. 获取发起调用的函数名，文件名，行号
	funcName, file, line, _ := runtime.Caller(1)

	// 3. 根据 ErrorLogFormat 组装成字符串
	logStr := fmt.Sprintf(ErrorLogFormat, currentTimeStr, component, runtime.FuncForPC(funcName).Name(), file, line, msg)
	// 4. 将字符串写入到文件中
	globalLogFile.WriteString(logStr)

	if ifDirectPrint {
		// 5. 将字符串打印到控制台
		color.Yellow(logStr)
	}
}

// Debug日志
func DebugLog(component string, msg string) {
	// 如果是debug模式，才会打印日志
	if !ifDebug {
		return
	}

	// 1. 获取当前的时间
	t := time.Now()
	currentTimeStr := t.Format("2006-01-02 15:04:05")
	// 2. 获取发起调用的函数名，文件名，行号
	funcName, file, line, _ := runtime.Caller(1)

	// 3. 根据 ErrorLogFormat 组装成字符串
	logStr := fmt.Sprintf(ErrorLogFormat, currentTimeStr, component, runtime.FuncForPC(funcName).Name(), file, line, msg)
	// 4. 将字符串写入到文件中
	globalLogFile.WriteString(logStr)

	if ifDirectPrint {
		// 5. 将字符串打印到控制台
		color.Blue(logStr)
	}
}

// Fatal日志
func FatalLog(component string, msg string) {
	// 1. 获取当前的时间
	t := time.Now()
	currentTimeStr := t.Format("2006-01-02 15:04:05")
	// 2. 获取发起调用的函数名，文件名，行号
	funcName, file, line, _ := runtime.Caller(1)

	// 3. 根据 ErrorLogFormat 组装成字符串
	logStr := fmt.Sprintf(ErrorLogFormat, currentTimeStr, component, runtime.FuncForPC(funcName).Name(), file, line, msg)
	// 4. 将字符串写入到文件中
	globalLogFile.WriteString(logStr)

	if ifDirectPrint {
		// 5. 将字符串打印到控制台
		color.Red(logStr)
	}

	// 6. 退出程序
	os.Exit(1)
}

// 关闭文件
func Close() {
	globalLogFile.Close()
}
