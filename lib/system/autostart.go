package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// SetAutoStart 设置当前应用开机自启动，支持传入额外启动参数
func SetAutoStart(args ...string) {
	appPath, err := os.Executable()
	if err != nil {
		fmt.Printf("获取应用路径失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("正在设置开机自启动...")
	if err := enableAutoStart(appPath, args...); err != nil {
		fmt.Printf("设置开机自启动失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("开机自启动设置成功")
}

// enableAutoStart 启用开机自启动，支持传入额外启动参数
func enableAutoStart(appPath string, args ...string) error {
	appName := strings.TrimSuffix(filepath.Base(appPath), filepath.Ext(appPath))
	serviceFile := fmt.Sprintf("/etc/systemd/system/%s.service", appName)

	// 构建带参数的启动命令
	execStart := appPath
	if len(args) > 0 {
		// 为每个参数添加引号以处理包含空格的情况
		quotedArgs := make([]string, len(args))
		for i, arg := range args {
			quotedArgs[i] = fmt.Sprintf("\"%s\"", strings.ReplaceAll(arg, "\"", "\\\""))
		}
		execStart += " " + strings.Join(quotedArgs, " ")
	}

	// 创建systemd服务文件内容
	content := fmt.Sprintf(`[Unit]
Description=%s Background Service
After=network.target

[Service]
Type=simple
ExecStart=%s
Restart=on-failure
RestartSec=10
User=root
WorkingDirectory=%s

[Install]
WantedBy=multi-user.target
`,
		appName,
		execStart,
		filepath.Dir(appPath))

	// 写入服务文件
	cmd := exec.Command("sudo", "bash", "-c",
		fmt.Sprintf("echo '%s' > %s", content, serviceFile))
	if err := cmd.Run(); err != nil {
		return err
	}

	// 重新加载systemd配置
	cmd = exec.Command("sudo", "systemctl", "daemon-reload")
	if err := cmd.Run(); err != nil {
		return err
	}

	// 启用服务
	cmd = exec.Command("sudo", "systemctl", "enable", appName+".service")
	return cmd.Run()
}
