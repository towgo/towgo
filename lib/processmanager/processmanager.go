/*
进程管理器
by:liangliangit
*/

package processmanager

import (
	"encoding/json"
	"errors"
	"os"
	"syscall"
	"time"

	"github.com/towgo/towgo/lib/system"
)

type processManagerJson struct {
	Pid int `json:"pid"`
}
type processManager struct {
	pid   int
	Error error
	path  string
}

var process processManager = processManager{}

func init() {
	process.pid = os.Getpid()
	process.path = system.GetPathOfProgram() + "/.processManager"
}

func (pm *processManager) getHandllerPid() (*processManagerJson, error) {
	processManagerJson := processManagerJson{}
	f, err := os.ReadFile(pm.path)
	if err != nil {
		return &processManagerJson, err
	}

	err = json.Unmarshal(f, &processManagerJson)
	if err != nil {
		return &processManagerJson, err
	}
	return &processManagerJson, nil
}
func (pmj *processManagerJson) isRuning() bool {

	if err := syscall.Kill(pmj.Pid, 0); err == nil {
		return true
	}

	return false
}

func fileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}

func (pm *processManager) write(pmj *processManagerJson) error {
	if fileExist(pm.path) {
		f, err := os.OpenFile(pm.path, os.O_WRONLY|os.O_TRUNC, 0600)
		if err != nil {
			return err
		}
		defer f.Close()
		jsonBytes, err := json.Marshal(pmj)
		if err != nil {
			return err
		}
		_, err = f.Write(jsonBytes)
		if err != nil {
			return err
		}
	} else {
		f, err := os.Create(pm.path)
		if err != nil {
			return err
		}
		defer f.Close()
		jsonBytes, err := json.Marshal(pmj)
		if err != nil {
			return err
		}
		_, err = f.Write(jsonBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func (pm *processManager) Start() bool {
	//检查是否已经有维护的程序
	pmj, err := pm.getHandllerPid()
	if err == nil {
		if pmj.isRuning() {
			pm.Error = errors.New("程序运行中:无法再次启动")
			return false
		}
	}

	//写入pid信息
	pmj.Pid = pm.pid
	err = pm.write(pmj)
	if err != nil {
		pm.Error = err
		return false
	}

	return true
}

func (pm *processManager) Stop() bool {
	pmj, err := pm.getHandllerPid()
	if err == nil {
		if pmj.isRuning() {
			syscall.Kill(pmj.Pid, syscall.SIGKILL)
			return true
		}
	}
	return false
}

func (pm *processManager) ReStart() bool {
	pmj, err := pm.getHandllerPid()
	if err == nil {
		if pmj.isRuning() {
			syscall.Kill(pmj.Pid, syscall.SIGKILL)
			time.Sleep(1 * time.Second)
			if pm.Start() {
				return true
			} else {
				return false
			}
		}
	}
	if pm.Start() {
		return true
	}
	return false
}

func GetManager() *processManager {
	return &process
}
