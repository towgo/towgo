package utils

import (
	"regexp"
	"strings"

	"path/filepath"

	"github.com/google/uuid"
	"github.com/towgo/towgo/lib/system"
)

func GetUuid() string {
	u4 := uuid.New()
	uuid := strings.Split(u4.String(), "-")[4]
	return uuid
}

type MicrosoftAuthenticateConfig struct {
	Domain map[string]string `json:"domain"`
	IsTLS  bool              `json:"is_tls"`
}

func GetMicrosoftAuthenticateConfig() MicrosoftAuthenticateConfig {
	basePath := system.GetPathOfProgram()
	var config MicrosoftAuthenticateConfig
	system.ScanConfigJson(filepath.Join(basePath, "config", "microsoft_login_config.json"), &config)
	return config
}

func GetCurrentApDomain(username string, config MicrosoftAuthenticateConfig) string {
	userList := strings.Split(username, "\\")
	v, ok := config.Domain[userList[0]]
	if !ok {
		return ""
	}
	return v
}
func VerifyEmailFormat(email string) bool {
	pattern := `\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*` //匹配电子邮箱
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(email)
}
