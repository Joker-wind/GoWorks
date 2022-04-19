package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

/*
nohup /opt/pin/ipfsPin &
*/
func main() {
	clock(60)
}

func clock(n int) {
	// n:定时间隔  返回ticker对象
	ticker := time.NewTicker(time.Second * time.Duration(n))
	sum := 0
	for range ticker.C {
		sum++
		saveLog(fmt.Sprintf("执行次数：%v", sum))
		data := getList()
		batchShell(data)
	}
}

func batchShell(d []string) {
	// 批量执行添加命令
	for _, cid := range d {
		//execCommand(string(cid))
		shell := fmt.Sprintf("crust tools ipfs pin add %v", cid)
		res := execCommand(shell)
		status := strings.Contains(res, "recursively")
		if status {
			// 如果成功就上报
			saveLog(fmt.Sprintf("添加成功，CID：%v", cid))
			innerIp := execCommand("/sbin/ifconfig -a|grep inet|grep -vE '127.0.0.1|inet6|172'|awk '{print $2}'|tr -d \"addr:\" | uniq")
			syncCid(cid, innerIp)
		}
	}
}

func saveLog(s string) {
	// 记录log日志
	tm := time.Now().Format("2006-01-02 15:04:05")
	f, err := os.OpenFile("log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
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

func execCommand(c string) string {
	// 执行命令并返回结果
	cmd := exec.Command("/bin/bash", "-c", c)

	stdout, _ := cmd.StdoutPipe()
	saveLog("开始执行命令：" + c)
	if err := cmd.Start(); err != nil {
		msg := "开始执行失败：" + err.Error()
		fmt.Println(msg)
		saveLog(msg)
		return ""
	}

	out, _ := ioutil.ReadAll(stdout)
	defer func() {
		if err := stdout.Close(); err != nil {
			return
		}
	}()

	if err := cmd.Wait(); err != nil {
		msg := "执行等待失败：" + err.Error()
		fmt.Println(msg)
		saveLog(msg)
		return ""
	}
	sOut := strings.TrimSuffix(string(out), "\n")
	saveLog("执行结果：" + sOut)
	return sOut
}

type FileList struct {
	Success bool        `json:"success"`
	Code    string      `json:"code"`
	Message interface{} `json:"message"`
	Data    []string    `json:"data"`
}

func getList() []string {
	// 获取列表
	client := &http.Client{
		Timeout: time.Second * 3,
	}
	url := "https://www.infin.cloud/web/user-file/not-ping-file-cid-list"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		msg := "获取列表，请求失败"
		saveLog(fmt.Sprintf("%v", err))
		saveLog(msg)
		fmt.Println(msg)
		return nil
	}
	saveLog(fmt.Sprintf("获取列表，请求成功"))
	fmt.Println("获取列表，请求成功")
	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		msg := "获取列表，解析响应失败"
		saveLog(fmt.Sprintf("%v", err))
		saveLog(msg)
		fmt.Println(msg)
		return nil
	}
	var lst FileList
	err = json.Unmarshal([]byte(string(body)), &lst)
	if err != nil {
		msg := "获取列表，解析参数失败"
		fmt.Println(msg)
		saveLog(msg)
		return nil
	}
	if lst.Success == true {
		msg := "获取列表，解析参数成功"
		saveLog(msg)
		fmt.Println(msg)
	}
	saveLog(fmt.Sprintf("%v", lst.Data))
	return lst.Data
}

func syncCid(c string, ip string) {
	// 上报CID
	client := &http.Client{
		Timeout: time.Second * 3,
	}
	fmt.Println(ip)
	url := fmt.Sprintf("https://www.infin.cloud/web/user-file/local-ping-sync?cid=%v&ip=%v", c, ip)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		msg := "上报CID，req构建失败"
		saveLog(fmt.Sprintf("%v", err))
		saveLog(msg)
		fmt.Println(msg)
		return
	}
	resp, err := client.Do(req)
	if err != nil {
		msg := "上报CID，请求失败"
		saveLog(fmt.Sprintf("%v", err))
		saveLog(msg)
		fmt.Println(msg)
		return
	}
	body, _ := ioutil.ReadAll(resp.Body)
	saveLog("上报成功:" + string(body))
	fmt.Println(string(body))
	defer func() {
		if err := resp.Body.Close(); err != nil {
			return
		}
	}()
}
