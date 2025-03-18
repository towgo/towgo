package main

import (
	"fmt"
	"github.com/towgo/towgo/lib/system"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var appName string = "oss copy module"
var appVersion string = "1.0.0"

var basePath = system.GetPathOfProgram()

type FileCopyConfig struct {
	SourcePath string `json:"sourcePath"`
	TargetPath string `json:"targetPath"`
	Speed      int    `json:"speed"`
}

func start() {
	var config FileCopyConfig
	system.ScanConfigJson(basePath+"/config/filecopy.json", &config)
	log.Println("源路径", config.SourcePath)
	log.Println("目标路径", config.TargetPath)
	log.Println("速率", config.Speed)

	for {
		log.Println("开始拷贝")
		err := CopyDir(config.SourcePath, config.TargetPath)
		if err != nil {
			log.Println("CopyDir err : ", err)
			return
		}
		log.Println("拷贝结束,下一次", config.Speed, "s后")
		time.Sleep(time.Second * time.Duration(config.Speed))
	}

}
func main() {
	//var wg sync.WaitGroup
	start()
	//wg.Wait()

}

// CopyDir 用于复制整个文件夹及其内容到目标文件夹，遇到重复文件跳过
func CopyDir(source string, target string) error {
	// 获取源文件夹的信息
	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}
	// 判断源是否为文件夹
	if !sourceInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", source)
	}

	// 创建目标文件夹（如果不存在）
	err = os.MkdirAll(target, sourceInfo.Mode())
	if err != nil {
		return err
	}

	var entries []os.DirEntry
	// 读取源文件夹下的所有文件和子文件夹
	entries, err = os.ReadDir(source)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(source, entry.Name())
		targetPath := filepath.Join(target, entry.Name())

		if entry.IsDir() {
			// 如果是子文件夹，递归调用CopyDir进行复制
			err = CopyDir(sourcePath, targetPath)
			if err != nil {
				return err
			}
		} else {
			// 检查目标文件是否已存在，如果存在则跳过复制
			if _, err := os.Stat(targetPath); err == nil {
				log.Println("文件已存在跳过")
				continue
			}
			// 如果是文件，进行文件复制操作
			err = CopyFile(sourcePath, targetPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// CopyFile 用于复制单个文件
func CopyFile(source string, target string) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	targetFile, err := os.Create(target)
	if err != nil {
		return err
	}
	defer targetFile.Close()

	_, err = io.Copy(targetFile, sourceFile)
	return err
}
