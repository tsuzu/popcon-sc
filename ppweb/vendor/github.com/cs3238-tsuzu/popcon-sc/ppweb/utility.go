package main

import (
	"fmt"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"math/rand"

	"golang.org/x/crypto/ssh/terminal"
)

func createWrapForm(req *http.Request) func(str string) int64 {
	return func(str string) int64 {
		arr, has := req.Form[str]
		if has && len(arr) != 0 {
			val, err := strconv.ParseInt(arr[0], 10, 64)

			if err != nil {
				return -1
			}
			return val
		}
		return -1
	}
}

func createWrapFormStr(req *http.Request) func(str string) string {
	return func(str string) string {
		arr, has := req.Form[str]
		if has && len(arr) != 0 {
			return arr[0]
		}
		return ""
	}
}

func CreateDefaultAdminUser() bool {
	fmt.Println("No user found in the DB")
	fmt.Println("You need to create the default admin")
	var id, name, email, pass, pass2 string

	fmt.Print("User ID: ")
	_, err := fmt.Scan(&id)

	if len(id) == 0 || err != nil {
		return false
	}

	fmt.Print("User Name: ")
	_, err = fmt.Scan(&name)

	if len(name) == 0 || err != nil {
		return false
	}

	fmt.Print("Email: ")
	_, err = fmt.Scan(&email)

	if len(email) == 0 || err != nil {
		return false
	}

	fmt.Print("Password (hidden): ")
	passArr, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return false
	}
	fmt.Println()

	fmt.Print("Password (confirmation): ")
	passArr2, err := terminal.ReadPassword(int(syscall.Stdin))

	if err != nil {
		return false
	}
	fmt.Println()

	pass = string(passArr)
	pass2 = string(passArr2)

	if pass != pass2 {
		fmt.Println("Different password")

		return false
	}

	_, err = mainDB.UserAdd(id, name, pass, email, 0)

	if err != nil {
		fmt.Println("Failed to create user. (", err.Error(), ")")

		return false
	}

	return true
}

func generateRandomString(strlen int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, strlen)
	for i := 0; i < strlen; i++ {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}

func CreateAdminUserAutomatically() bool {
	pass := generateRandomString(8)
	_, err := mainDB.UserAdd("admin", "admin", pass, "admin_null", 0)

	if err == nil {
		DBLog.Infof("Default User(ID: admin, Pass: %s)", pass)

		return true
	}

	return false
}
