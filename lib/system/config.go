package system

import (
	"encoding/json"
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
