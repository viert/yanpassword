package crypter

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

const (
	authFilename = ".yanpasswd_auth"
)

// AuthData represents Yandex user auth data to authenticate with in Webdav service
type AuthData struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (ad *AuthData) dump() ([]byte, error) {
	return json.Marshal(ad)
}

func readPassword() (string, error) {
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(bytePassword), nil
}

// GetMasterPassword gets a masterpassword from user input
func GetMasterPassword() (string, error) {
	var err error
	var pwd string

	for pwd == "" {
		fmt.Print("Enter Master Password: ")
		pwd, err = readPassword()
		if err != nil {
			return "", err
		}
		pwd = strings.TrimSpace(pwd)
	}
	return pwd, nil
}

func getAuthDataFilename() string {
	return path.Join(os.Getenv("HOME"), authFilename)
}

// ReadAuthData reads auth data from a crypted file located in ~/.yanpasswd_auth
func ReadAuthData(masterPassword string) (*AuthData, error) {
	fileName := getAuthDataFilename()
	st, err := os.Stat(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("Yandex auth data file %s not found, creating new...\n", fileName)

			authData, err := inputAuthData()
			if err != nil {
				return nil, err
			}

			err = saveAuthData(authData, masterPassword)
			if err != nil {
				return nil, err
			}
			return authData, nil
		}
		return nil, err
	}

	perm := st.Mode().Perm()
	if perm&63 > 0 {
		return nil, fmt.Errorf("Yandex auth data file is accessible to other users or groups")
	}

	authData, err := loadAuthData(masterPassword)
	if err != nil {
		return nil, err
	}

	return authData, nil
}

func inputAuthData() (*AuthData, error) {
	var username string
	var password string
	var rd *bufio.Reader
	var err error

	rd = bufio.NewReader(os.Stdin)
	for username == "" {
		fmt.Print("Yandex Webdav username: ")
		username, err = rd.ReadString('\n')
		if err != nil {
			return nil, err
		}
		username = strings.TrimSpace(username)
	}
	for password == "" {
		fmt.Print("Yandex Webdav password: ")
		password, err = readPassword()
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(password)
	}

	return &AuthData{username, password}, nil
}

func saveAuthData(data *AuthData, masterPassword string) error {
	fmt.Println("Saving auth data...")
	fileName := getAuthDataFilename()
	byteData, err := data.dump()
	encrypted, err := Encrypt(byteData, masterPassword)
	if err != nil {
		fmt.Printf("Error encrypting data: %s\n", err)
		return err
	}
	ioutil.WriteFile(fileName, encrypted, os.FileMode(0600))
	return nil
}

func loadAuthData(masterPassword string) (*AuthData, error) {
	var data AuthData

	fileName := getAuthDataFilename()
	f, err := os.Open(fileName)
	if err != nil {
		fmt.Printf("Error reading auth data file: %s\n", err)
		return nil, err
	}
	defer f.Close()
	encrypted, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Printf("Error reading auth data file: %s\n", err)
		return nil, err
	}

	byteData, err := Decrypt(encrypted, masterPassword)
	if err != nil {
		fmt.Printf("Error decrypting data file: %s\n", err)
		return nil, err
	}

	err = json.Unmarshal(byteData, &data)
	if err != nil {
		fmt.Printf("Error unmarshalling auth data: %s\n", err)
		return nil, err
	}
	return &data, nil
}
