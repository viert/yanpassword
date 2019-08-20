package manager

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/viert/yanpassword/client"
	"github.com/viert/yanpassword/crypter"
	"github.com/viert/yanpassword/term"
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

func getAuthDataFilename() string {
	return path.Join(os.Getenv("HOME"), authFilename)
}

func authdataExists() bool {
	_, err := os.Stat(getAuthDataFilename())
	if err == nil {
		return true
	}
	return !os.IsNotExist(err)
}

func (m *Manager) acquireAuthData() error {
	if authdataExists() {
		for {
			pwd, err := m.getMasterPassword()
			if err != nil {
				return err
			}
			m.masterPassword = pwd

			fmt.Println("Checking Yandex webdav auth...")
			authData, err := m.loadWebdavAuth()
			if err != nil {
				continue
			}

			if m.checkWebdavAuth(authData) {
				m.webdavAuthData = authData
				return nil
			}

			fmt.Println(`
Looks like your current auth data file contains invalid authentication data.
To prevent data loss I'm going to give up now. If you want to create a new 
auth data file, please remove ~/.yanpasswd_auth manually and restart yanpassword.` + "\n")
			return fmt.Errorf("Valid auth data file contains invalid credentials")

		}
	} else {
		fmt.Println(`
Seems like you're running Yanpassword for the first time.
Let's set your master password. If you already have yanpassword data on Yandex.Disk, feel free
to use the same password as before or create a new one - doesn't matter. If anything goes wrong
with decrypting your existing data I'll prompt for a proper password.` + "\n")
		// Set new MP
		pwd, err := m.setNewMasterPassword()
		if err != nil {
			return err
		}
		m.masterPassword = pwd
		return m.createNewAuthData()
	}
}

func (m *Manager) createNewAuthData() error {
	var authData AuthData
	var err error
	fmt.Println(`
Let's deal with your Yandex.Disk account. 
You can use your primary Yandex account password, however, it's recommended 
to turn on application passwords at https://passport.yandex.ru and create
a special password for Yanpassword only (use Yandex.Disk/Webdav type of password).` + "\n")
	for {
		authData, err = m.inputWebdavAuth()
		if err != nil {
			return err
		}
		if m.checkWebdavAuth(authData) {
			m.webdavAuthData = authData
			break
		}
	}
	return m.saveWevdavAuth(authData)
}

func (m *Manager) setNewMasterPassword() (string, error) {
	var pwd []byte
	var pwdConfirm []byte
	var err error

	for {
		pwd, err = m.rl.ReadPassword("Set Master Password: ")
		if err != nil {
			return "", err
		}

		if string(pwd) == "" {
			term.Errorf("Master password can't be empty\n\n")
			continue
		}

		pwdConfirm, err = m.rl.ReadPassword("Confirm Master Password: ")
		if err != nil {
			return "", err
		}

		if string(pwd) == string(pwdConfirm) {
			break
		}
		term.Errorf("Passwords don't match\n\n")
	}
	return string(pwd), nil
}

func (m *Manager) getMasterPassword() (string, error) {
	for {
		pwd, err := m.rl.ReadPassword("Enter Master Password: ")
		if err != nil {
			return "", err
		}

		if string(pwd) != "" {
			return string(pwd), nil
		}
	}
}

func (m *Manager) inputWebdavAuth() (AuthData, error) {
	var ad AuthData
	var username string
	var password []byte
	var err error

	rd := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Yandex webdav username: ")
		username, err = rd.ReadString('\n')
		if err != nil {
			return ad, err
		}
		username = strings.TrimSpace(username)
		if username != "" {
			break
		}
		term.Errorf("Username can't be empty\n")
	}

	password, err = m.rl.ReadPassword("Yandex webdav password: ")
	if err != nil {
		return ad, err
	}

	ad.Username = username
	ad.Password = string(password)

	return ad, nil
}

func (m *Manager) checkWebdavAuth(authData AuthData) bool {
	err := client.CheckAuth(authData.Username, authData.Password)
	if err != nil {
		term.Errorf("Authentication Error: %s\n", err)
	}
	return err == nil
}

func (m *Manager) saveWevdavAuth(authData AuthData) error {
	authJSON, err := authData.dump()
	if err != nil {
		term.Errorf("Error marshalling auth data, this must be a bug: %s\n", err)
		return err
	}

	data, err := crypter.Encrypt(authJSON, m.masterPassword)
	if err != nil {
		term.Errorf("Error encrypting auth data, this must be a bug: %s\n", err)
		return err
	}

	filename := getAuthDataFilename()
	err = ioutil.WriteFile(getAuthDataFilename(), data, os.FileMode(0600))
	if err != nil {
		term.Errorf("Error saving authentication file %s: %s\n", filename, err)
	} else {
		term.Successf("Authentication file saved successfully\n")
	}

	return err
}

func (m *Manager) loadWebdavAuth() (AuthData, error) {
	var ad AuthData

	f, err := os.Open(getAuthDataFilename())
	if err != nil {
		term.Errorf("Error opening auth data file: %s\n", err)
		return ad, err
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		term.Errorf("Error loading auth data file: %s\n", err)
		return ad, err
	}

	authJSON, err := crypter.Decrypt(data, m.masterPassword)
	if err != nil {
		term.Errorf("Error decrypting auth data file: %s\n", err)
		return ad, err
	}
	err = json.Unmarshal(authJSON, &ad)
	if err != nil {
		term.Errorf("Error unmarshalling auth data file: %s\n", err)
	}
	return ad, err

}
