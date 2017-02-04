package main

import (
	"fmt"
	"os/exec"
)

func SendMail(to string, subject string, body string) error {
	cmdArr := make([]string, 0, len(settingManager.Get().SendmailCommand))

	for _, s := range settingManager.Get().SendmailCommand {
		switch s {
		case "#{to}":
			cmdArr = append(cmdArr, to)
		case "#{subject}":
			cmdArr = append(cmdArr, subject)
		case "#{body}":
			cmdArr = append(cmdArr, body)
		default:
			cmdArr = append(cmdArr, s)
		}
	}

	b, err := exec.Command(cmdArr[0], cmdArr[1:]...).CombinedOutput()

	if err != nil {
		return fmt.Errorf("output: %s, error: %s", string(b), err.Error())
	}
	MailLog.WithField("to", to).WithField("subject", subject).Info("A mail was sent successfully")

	return nil
}
