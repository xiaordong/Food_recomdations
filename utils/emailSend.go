package utils

import (
	"crypto/tls"
	gomail "gopkg.in/mail.v2"
	"log"
	"strconv"
)

func SendEmailCode(email, data, code string) (string, error) {
	m := gomail.NewMessage()
	m.SetHeader("From", "2132047479@qq.com")
	m.SetHeader("To", email)
	m.SetHeader("Subject", "验证码")
	m.SetBody("text/html", data)
	port, _ := strconv.Atoi("587")
	d := gomail.NewDialer("smtp.qq.com", port, "2132047479@qq.com", "rmubxprbsgmcehia")
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	err := d.DialAndSend(m)
	if err != nil {
		log.Println("发送邮件时出现错误:", err)
		return "", err
	}
	return code, nil
}
