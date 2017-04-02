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

const MAIL_MINIMUM_INTERVAL = 10 * time.Minute

var ErrMailWasSent = errors.New("The mail was sent recently, try later.")

func MailSendConfirmUser(iid int64, userName, email string) error {
	if ok, err := mainRM.TokenExists(MAILLASTSENDSERVICE, email); err != nil {
		DBLog().WithError(err).WithField("email", email).Error("TokenExists error")
	} else if ok {
		return ErrMailWasSent
	}

	if err := mainRM.TokenRegister(MAILLASTSENDSERVICE, email, MAIL_MINIMUM_INTERVAL); err != nil {
		DBLog().WithError(err).WithField("email", email).Error("TokenRegister error")
	}

	token, err := mainRM.TokenGenerateAndRegisterWithValue(MAILCONFTOKENSERVICE, time.Duration(settingManager.Get().MailConfTokenExpirationInMinutes)*time.Minute, iid)

	if err != nil {
		DBLog().WithError(err).Error("Token generation for mail confirmation failed")

		return err
	}

	s, b := MailCreateConfirmUser(userName, path.Join(settingManager.Get().PublicHost, "/signup/account_confirm?token="+token))
	err = SendMail(email, s, b)

	if err != nil {
		MailLog().WithError(err).Error("Sending mail failed")

		return err
	}

	return nil
}
