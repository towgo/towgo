package mail

import (
	"errors"
	"io"
	"net/http"
	"net/smtp"
	"path/filepath"
	"sync"

	"github.com/jordan-wright/email"
	"github.com/towgo/towgo/lib/system"
)

var mail *Mail

type Mail struct {
	sync.Mutex
	Account     string
	Password    string
	From        string
	Server      string
	Port        string
	MailTo      []string
	SubJect     string
	Body        string
	Reader      io.Reader
	Filename    string
	ContentType string
}

type Config struct {
	MailAccount  string
	MailPassword string
	MailServer   string
	MailPort     string
	MailFrom     string
}

func Init(config Config) {
	mail = &Mail{}
	mail.Account = config.MailAccount
	mail.Password = config.MailPassword
	mail.Server = config.MailServer
	mail.Port = config.MailPort
	mail.From = config.MailFrom
}

func InitTestApi() {
	http.HandleFunc("/mail/test", func(w http.ResponseWriter, r *http.Request) {
		err := Test()
		if err != nil {
			w.Write([]byte(err.Error()))
		} else {
			w.Write([]byte("邮件发送成功"))
		}
	})
}

func Send(mailTo []string, subject, body string) error {
	mail.Lock()
	defer mail.Unlock()
	if mail.Account == "" {
		return errors.New("账户未初始化")
	}
	mail.MailTo = mailTo
	mail.SubJect = subject
	mail.Body = body
	return mail.Send()
}

func Test() error {
	basePath := system.GetPathOfProgram()
	var config struct {
		NoticeAddress []string
	}

	system.ScanConfigJson(filepath.Join(basePath, "/config/mail.config.json"), &config)
	body := "<body>"
	body = body + "<p>邮箱检测</p>"
	body = body + "</body>"
	return Send(config.NoticeAddress, "邮箱检测", body)
}

func (m *Mail) Send() error {
	em := email.NewEmail()

	em.From = m.From
	em.To = m.MailTo
	em.Subject = m.SubJect
	if m.Reader != nil {
		em.Attach(m.Reader, m.Filename, m.ContentType)
	}

	em.HTML = []byte(m.Body)
	err := em.Send(m.Server+":"+m.Port, smtp.PlainAuth(m.Account, m.Account, m.Password, m.Server))
	if err != nil {
		return err
	}
	return nil
}
