package smtps

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

//Message ...
type Message struct {
	senderID   string
	recipients []string
	subject    string
	body       string
}

//Server ...
type Server struct {
	credentials Credentials
	host        string
	port        uint
	tls         *tls.Config
}

//Credentials ...
type Credentials struct {
	Pass  string
	Login string
}

//NewServer ...
func NewServer(host string, port uint, cred Credentials) *Server {
	return &Server{
		cred,
		host,
		port,
		&tls.Config{
			InsecureSkipVerify: true,
			ServerName:         host,
		},
	}
}

//Addr...
func (s *Server) Addr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

//Send message
func (s *Server) Send(m *Message) error {
	auth := smtp.PlainAuth("", s.credentials.Login, s.credentials.Pass, s.host)
	conn, err := tls.Dial("tcp", s.Addr(), s.tls)
	if err != nil {
		return err
	}

	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return err
	}
	// step 1: Use Auth
	if err = client.Auth(auth); err != nil {
		return err
	}

	// step 2: add all from and to
	if err = client.Mail(s.credentials.Login); err != nil {
		return err
	}
	for _, k := range m.recipients {
		if err = client.Rcpt(k); err != nil {
			return err
		}
	}
	// Data
	w, err := client.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(m.Build()))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	err = client.Quit()
	if err != nil {
		return err
	}
	return nil
}

//Mail ...
func Mail() *Message {
	return &Message{}
}

//From ...
func (m *Message) From(nick, mail string) *Message {
	m.senderID = nick + "<" + mail + ">"
	return m
}

//Recipients ...
func (m *Message) Recipients(recips []string) *Message {
	m.recipients = recips
	return m
}

//Subject ...
func (m *Message) Subject(sub string) *Message {
	m.subject = sub
	return m
}

//Body ...
func (m *Message) Body(body string) *Message {
	m.body = body
	return m
}

//Build ...
func (m *Message) Build() string {
	message := "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n"
	message += fmt.Sprintf("From: %s\r\n", m.senderID)
	if len(m.recipients) > 0 {
		message += fmt.Sprintf("To: %s\r\n", strings.Join(m.recipients, ";"))
	}
	message += fmt.Sprintf("Subject: %s\r\n", m.subject)
	message += "\r\n" + m.body
	return message
}
