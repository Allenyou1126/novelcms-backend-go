package utils

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/smtp"
	"os"
	"strconv"
)

var auth smtp.Auth
var server string
var serverAddr string
var senderMail string
var tlsEnabled bool

func init() {
	serverPort, err := strconv.Atoi(GetConfigWithDefault("SMTP_SERVER_PORT", "25"))
	if err != nil {
		slog.Error("Invalid smtp server port.")
		os.Exit(1)
	}
	userName := GetConfigWithDefault("SMTP_USERNAME", "")
	password := GetConfigWithDefault("SMTP_PASSWORD", "")
	senderMail = GetConfigWithDefault("SMTP_MAIL", "test@example.com")
	serverAddr = GetConfigWithDefault("SMTP_SERVER_ADDR", "smtp.example.com")
	auth = smtp.PlainAuth("", userName, password, serverAddr)
	server = fmt.Sprintf("%s:%d", serverAddr, serverPort)
	tlsEnabled = func() bool {
		ret := GetConfigWithDefault("SMTP_TLS", "false")
		if ret == "true" || ret == "TRUE" || ret == "True" {
			return true
		}
		return false
	}()
}

func encode(content string) string {
	return "=?UTF-8?B?" + base64.StdEncoding.EncodeToString([]byte(content)) + "?="
}

func SendMail(target string, title string, content string) error {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("From: %s\r\n", senderMail))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", target))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", encode(title)))
	buf.WriteString("Content-Type: text/html;charset=utf-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(content)
	targets := append(make([]string, 1), target)
	if !tlsEnabled {
		return smtp.SendMail(server, auth, senderMail, targets, buf.Bytes())
	}
	conn, err := tls.Dial("tcp", server, &tls.Config{
		ServerName: serverAddr,
	})
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, serverAddr)
	if err != nil {
		return err
	}
	if err = client.Auth(auth); err != nil {
		return err
	}
	if err = client.Mail(senderMail); err != nil {
		return err
	}

	if err = client.Rcpt(target); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}
	_ = client.Quit()
	return nil
}
