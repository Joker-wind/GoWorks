package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func main() {
	file:="test.txt"
	ReadIp(file)
}

func ReadIp(fileName string) []string {
	f,err:=os.Open(fileName)
	if err != nil {
		fmt.Println(err)
		return []string{}
	}
	defer func() {
		if err = f.Close(); err != nil {
			return
		}
	}()
	fd,err:=ioutil.ReadAll(f)
	if err!=nil{
		fmt.Println(err)
		return []string{}
	}
	fmt.Println(string(fd))

	return []string{}
}

