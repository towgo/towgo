package system

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func ScanConfigJson(path string, dest any) {
	//只读打开
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		log.Print(err.Error())
		return
	}
	defer file.Close()
	b, err := io.ReadAll(file)
	if err != nil {
		log.Print(err.Error())
		return
	}
	err = json.Unmarshal(b, dest)
	if err != nil {
		log.Print(err.Error())
		return
	}
}

func SaveConfigJson(path string, dest any) error {
	// 将结构体转换为 JSON 格式的字节流
	jsonData, err := json.Marshal(dest)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	// 将 JSON 数据写入到本地文件
	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		fmt.Println("Error writing JSON to file:", err)
		return err
	}
	return nil
}
