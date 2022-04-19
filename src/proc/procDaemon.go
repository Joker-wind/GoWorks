package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"
)

// 设置
var LogPath = "/opt/procDaemon/log"
var PidPath = "/opt/procDaemon/pid"
var workPath = "/opt/procDaemon"
var wg sync.WaitGroup

func init() {
	if _, err := os.Stat(workPath); err != nil {
		err := os.MkdirAll(workPath, 0644)
		if err != nil {
			saveLog("目录创建错误：" + err.Error())
			os.Exit(1)
		}
	}
}

func main() {
	var msg string
	if len(os.Args) != 2 {
		msg = "参数错误，请重新输入"
		fmt.Println(msg)
		return
	}
	switch os.Args[1] {
	case "start":
		pid := getPid()
		if pid != "" && pid != "0" {
			return
		}
		cmd := Daemon()
		if cmd == nil {
			log.Println("服务启动失败！")
			return
		}
	case "status":
		pid, _ := strconv.Atoi(getPid())
		proc, err := os.FindProcess(pid)
		if err != nil || pid == 0 {
			fmt.Println("procDaemon 未启动")
			return
		}
		if proc != nil {
			fmt.Println("procDaemon 已启动")
			return
		}
	case "stop":
		msg = "procDaemon 已退出！"
		pid, _ := strconv.Atoi(getPid())
		proc, err := os.FindProcess(pid)
		if err != nil || pid == 0 {
			fmt.Println("procDaemon 未启动")
			return
		}
		err = proc.Kill()
		if err != nil {
			fmt.Println("procDaemon 退出失败")
			return
		}
		setPid("0")
		fmt.Println(msg)
		saveLog(msg)
	case "--daemon":
		// 后台的业务代码
		wg.Add(1)
		ch1 := make(chan int, 1)
		go clock(2, ch1)
		for i := range ch1 {
			saveLog(fmt.Sprintf("执行次数：%v", i))
		}
		wg.Wait()
	}
}

func getPid() string {
	pidFile, err := os.OpenFile(PidPath, os.O_CREATE|os.O_RDONLY, 0444)
	if err != nil {
		saveLog("获取PID出现错误：" + err.Error())
		return ""
	}
	defer func() {
		if err = pidFile.Close(); err != nil {
			return
		}
	}()
	pid, err := ioutil.ReadAll(pidFile)
	if err != nil {
		saveLog("转换PID出现错误：" + err.Error())
		return ""
	}
	return string(pid)
}

func setPid(s string) {
	pidFile, err := os.OpenFile(PidPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0444)
	if err != nil {
		saveLog("设置PID出现错误：" + err.Error())
		return
	}
	defer func() {
		if err = pidFile.Close(); err != nil {
			return
		}
	}()
	_, err = pidFile.WriteString(s)
	if err != nil {
		return
	}
}

func Daemon() *exec.Cmd {
	if os.Getppid() != 1 {
		var msg string
		filepath := os.Args[0]
		cmd := exec.Command(filepath, "--daemon")
		if err := cmd.Start(); err != nil {
			msg = "服务启动失败：" + err.Error()
			saveLog(msg)
			return nil
		}
		msg = fmt.Sprintf("服务pid=%d:开始运行...", cmd.Process.Pid)
		saveLog(msg)
		pidStr := strconv.Itoa(cmd.Process.Pid)
		setPid(pidStr)
		fmt.Println("procDaemon 已启动！")
		os.Exit(0)
		return cmd
	}
	return nil
}

func clock(n int, c chan<- int) {
	// n:定时间隔
	defer wg.Done()
	ticker := time.NewTicker(time.Second * time.Duration(n))
	sum := 0
	for range ticker.C {
		sum++
		c <- sum
	}
}

func saveLog(s string) {
	// 记录log日志
	tm := time.Now().Format("2006-01-02 15:04:05")
	f, err := os.OpenFile(LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	defer func() {
		if err = f.Close(); err != nil {
			return
		}
	}()
	if err != nil {
		return
	}
	str := tm + "\t" + s + "\n"
	_, err = f.WriteString(str)
	if err != nil {
		return
	}
}
