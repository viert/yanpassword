package manager

import (
	"client"
	"crypter"
	"encoding/json"
	"fmt"
	"term"
)

func (m *Manager) acquirePassdb() error {
	var data []byte
	var decrypted []byte
	var err error

	cli := client.NewPassdbClient(m.webdavAuthData.Username, m.webdavAuthData.Password)
	fmt.Println("Loading remote data...")
	data, err = cli.Load()
	if err != nil {
		if client.Is404(err) {
			// data doesn't exist
			fmt.Println("No remote data found, creating passdb from scratch")
			m.data = m.createPassdb()
			return nil
		}
		return err
	}

	// data exists
	passwd := m.masterPassword
	for {
		decrypted, err = crypter.Decrypt(data, passwd)
		if err == nil {
			break
		}

		term.Errorf("Error decrypting yanpasword data. Master password's changed?\n")
		bytePwd, err := m.rl.ReadPassword("Previous Master Password: ")
		if err != nil {
			return err
		}
		passwd = string(bytePwd)
	}

	err = json.Unmarshal(decrypted, &m.data)
	if err != nil {
		term.Errorf("Error unmarshalling yanpassword data: %s\n", err)
		return err
	}

	term.Successf("Remote data loaded and parsed. %d items in total.\n", len(m.data))
	return nil
}

func (m *Manager) createPassdb() serviceData {
	return make(serviceData)
}

func (m *Manager) savePassdb() error {
	data, err := json.Marshal(m.data)
	if err != nil {
		term.Errorf("Error marshalling yanpassword data: %s\n", err)
		return err
	}

	encrypted, err := crypter.Encrypt(data, m.masterPassword)
	if err != nil {
		term.Errorf("Error encrypting yanpassword data: %s\n", err)
		return err
	}

	cli := client.NewPassdbClient(m.webdavAuthData.Username, m.webdavAuthData.Password)
	return cli.Save(encrypted)
}
