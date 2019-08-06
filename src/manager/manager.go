package manager

import (
	"client"
	"crypter"
	"encoding/json"
	"fmt"
)

// Manager is the main exported class
type Manager struct {
	masterPassword string
	data           serviceData
}

type serviceInfo struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Comment   string `json:"comment"`
	UpdatedAt string `json:"updated_at"`
	URL       string `json:"url"`
}

type serviceData map[string]*serviceInfo

// NewManager creates and initializes a new manager instance
func NewManager() (*Manager, error) {
	pwd, err := crypter.GetMasterPassword()
	if err != nil {
		return nil, err
	}
	m := &Manager{
		masterPassword: pwd,
	}
	return m, nil
}

// Start function gets auth data and starts the manager
func (m *Manager) Start() error {
	authData, err := crypter.ReadAuthData(m.masterPassword)
	if err != nil {
		return err
	}

	err = client.CheckAuth(authData)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = m.loadData(authData)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) loadData(authData *crypter.AuthData) error {
	pdbc := client.NewPassdbClient(authData)
	encrypted, err := pdbc.Load()

	if err != nil {
		if client.Is404(err) {
			m.data = make(serviceData)
		} else {
			fmt.Printf("Error reading data file: %s\n", err)
			return err
		}
	} else {
		// decrypt data
		data, err := crypter.Decrypt(encrypted, m.masterPassword)
		if err != nil {
			fmt.Printf("Error decrypting data file: %s\n", err)
			return err
		}
		err = json.Unmarshal(data, &m.data)
		if err != nil {
			fmt.Printf("Error unmarshalling data file: %s\n", err)
			return err
		}
	}
	return nil
}
