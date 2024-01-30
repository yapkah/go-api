package mail

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"

	"github.com/yapkah/go-api/models"
	"github.com/yapkah/go-api/pkg/e"
)

//EmailAuth struct
type EmailAuth struct {
	Identity string
	Username string
	Password string
	Host     string
	Port     string
}

//SendMailData struct
type SendMailData struct {
	Subject string
	Message string
	Type    string
	// FromMail string
	FromName string
	ToEmail  []string
	ToName   []string
	CCEmail  []string
	CCName   []string
	BccEmail []string
}

//SendMail func
func (s SendMailData) SendMail(username, password, mailhost, mailport string) error {
	m := EmailAuth{
		Identity: "",
		Username: username,
		Password: password,
		Host:     mailhost,
		Port:     mailport,
	}
	if m.Username == "" || m.Password == "" || m.Host == "" || m.Port == "" {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "please_setup_email_auth_setting"}
	}

	auth := smtp.PlainAuth(
		m.Identity,
		m.Username,
		m.Password,
		m.Host,
	)

	if s.Subject == "" {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "send_mail_subject_invalid"}
	}

	if s.Message == "" {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "send_mail_message_invalid"}
	}

	// email
	var recipientMail []string
	var toHeaderMail []string
	var ccHeaderMail []string
	var bccHeaderMail []string

	// to email and to name checking
	if len(s.ToEmail) != len(s.ToName) {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "to_email_must_same_length_with_to_name"}
	}

	// to email process
	if len(s.ToEmail) != 0 {
		for index, to := range s.ToEmail {
			str := s.ToName[index] + " <" + to + ">"
			toHeaderMail = append(toHeaderMail, str)
			recipientMail = append(recipientMail, to)
		}
	}

	// cc email and cc name checking
	if len(s.CCEmail) != len(s.CCName) {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "cc_email_must_same_length_with_cc_name"}
	}

	// cc email process
	if len(s.CCEmail) != 0 {
		for index, cc := range s.CCEmail {
			str := s.CCName[index] + " <" + cc + ">"
			ccHeaderMail = append(ccHeaderMail, str)
			recipientMail = append(recipientMail, cc)
		}
	}

	// bcc email process
	if len(s.BccEmail) != 0 {
		for _, bcc := range s.BccEmail {
			bccHeaderMail = append(bccHeaderMail, bcc)
			recipientMail = append(recipientMail, bcc)
		}
	}

	if len(recipientMail) == 0 {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.ERROR, Msg: "there_are_no_receipient"}
	}

	fromName := ""
	if s.FromName != "" {
		fromName = s.FromName
	} else {
		// if setting.EmailSetting.MailName != "" {
		// 	fromName = setting.EmailSetting.MailName
		// }
		fromName = "No Reply"

	}

	toHeader := strings.Join(toHeaderMail, ",")
	ccHeader := strings.Join(ccHeaderMail, ",")
	bccHeader := strings.Join(bccHeaderMail, ",")

	header := make(map[string]string)
	header["MIME-Version"] = "1.0"
	header["Content-Transfer-Encoding"] = "base64"
	header["From"] = fromName + " <" + m.Username + ">"
	header["Subject"] = s.Subject

	if s.Type == "HTML" {
		header["Content-Type"] = "text/html; charset=\"utf-8\""
	} else {
		header["Content-Type"] = "text/plain; charset=\"utf-8\""
	}

	if toHeader != "" {
		header["To"] = toHeader
	}
	if ccHeader != "" {
		header["Cc"] = ccHeader
	}
	if bccHeader != "" {
		header["Bcc"] = bccHeader
	}

	msg := ""
	for k, v := range header {
		msg += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	msg += "\r\n" + base64.StdEncoding.EncodeToString([]byte(s.Message))

	addr := m.Host + ":" + m.Port

	err := smtp.SendMail(
		addr,          // server:port
		auth,          // auth
		m.Username,    // from email_address
		recipientMail, // to []email_address
		[]byte(msg),   // msg content_here
	)

	if err != nil {
		return &e.CustomError{HTTPCode: http.StatusUnprocessableEntity, Code: e.UNPROCESSABLE_ENTITY, Msg: err.Error(), Data: err}
	}

	// email log
	toEmail, _ := json.Marshal(s.ToEmail)
	data, _ := json.Marshal(s)

	emailLog := models.EmailLog{
		Email:    string(toEmail),
		Provider: "SMTP",
		Data:     string(data),
	}

	db := models.GetDB() // no need transaction because if failed no need rollback

	_, err = models.AddEmailLog(db, emailLog)
	if err != nil {
		return err
	}

	return nil
}
