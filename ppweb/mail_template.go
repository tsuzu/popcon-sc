package main

import (
	"bytes"
	"errors"
	"path"
	"text/template"
	"time"
)

func MailCreateConfirmUser(userName, confURL string) (string, string) {
	subject := "popconの登録を完了して下さい"

	body := "Thank you for registering popcon!\r\n\r\n{{.UserName}}様\r\npopconへの登録ありがとうございます。\r\n登録を完了するためには以下のリンクにアクセスしてください。\r\n{{.ConfURL}}\r\n\r\npopcon"

	var buf bytes.Buffer
	template.Must(template.New("").Parse(body)).Execute(&buf, map[string]string{"UserName": userName, "ConfURL": confURL})

	return subject, buf.String()
}

var MAILCONFTOKENSERVICE = "mail_conf"
var MAILLASTSENDSERVICE = "mail_ls"

var ErrMailWasSent = errors.New("The mail was sent recently, try later.")

func MailSendConfirmUser(iid int64, userName, email string) error {
	if ok, err := mainRM.TokenExists(MAILLASTSENDSERVICE, email); err != nil {
		DBLog().WithError(err).WithField("email", email).Error("TokenExists error")
	} else if ok {
		return ErrMailWasSent
	}

	interval, err := mainRM.MailMinInterval()

	if err != nil {
		DBLog().WithError(err).Error("MailMinInterval() error")

		return err
	}

	exp, err := mainRM.MailConfTokenExpiration()

	if err != nil {
		DBLog().WithError(err).Error("CSRFConfTokenExpiration() error")

		return err
	}

	if err := mainRM.TokenRegister(MAILLASTSENDSERVICE, email, time.Duration(interval)*time.Minute); err != nil {
		DBLog().WithError(err).WithField("email", email).Error("TokenRegister error")
	}

	ph, err := mainRM.PublicHost()

	if err != nil {
		DBLog().WithError(err).Error("PublicHost() error")

		return err
	}

	token, err := mainRM.TokenGenerateAndRegisterWithValue(MAILCONFTOKENSERVICE, time.Duration(exp)*time.Minute, iid)

	if err != nil {
		DBLog().WithError(err).Error("Token generation for mail confirmation failed")

		return err
	}

	s, b := MailCreateConfirmUser(userName, path.Join(ph, "/signup/account_confirm?token="+token))
	err = SendMail(email, s, b)

	if err != nil {
		MailLog().WithError(err).Error("Sending mail failed")

		return err
	}

	return nil
}
