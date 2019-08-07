package manager

import (
	"client"
	"crypter"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"term"

	"github.com/chzyer/readline"
)

type serviceData map[string]*serviceInfo
type cmdHandler func(string, string, ...string)

// Manager is the main exported class
type Manager struct {
	masterPassword string
	data           serviceData
	rl             *readline.Instance
	stopped        bool
	handlers       map[string]cmdHandler
	webdavAuthData AuthData
}

type serviceInfo struct {
	Name      string `json:"name"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Comment   string `json:"comment"`
	UpdatedAt string `json:"updated_at"`
	URL       string `json:"url"`
}

// NewManager creates and initializes a new manager instance
func NewManager() (*Manager, error) {
	m := new(Manager)
	m.setupHandlers()
	err := m.setupReadline()
	if err != nil {
		return nil, err
	}
	return m, nil
}

// Start function gets auth data and starts the manager
func (m *Manager) Start() error {

	err := m.acquireAuthData()
	if err != nil {
		return err
	}
	// authData, err := crypter.ReadAuthData(m.masterPassword)
	// if err != nil {
	// 	return err
	// }

	// err = client.CheckAuth(authData)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }

	// err = m.loadData(authData)
	// if err != nil {
	// 	return err
	// }

	m.setPrompt()
	m.cmdLoop()

	return nil
}

func (m *Manager) loadData(authData AuthData) error {
	pdbc := client.NewPassdbClient(authData.Username, authData.Password)
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
			// TODO: prompt for an optional remote-side password for the case
			// it doesn't match the current master password. Next time being saved
			// the data will be encrypted with the new one.

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

func (m *Manager) cmdLoop() {
	for !m.stopped {
		line, err := m.rl.Readline()
		if err == readline.ErrInterrupt {
			continue
		} else if err == io.EOF {
			m.stopped = true
		}
		m.cmd(line)
	}
}

func (m *Manager) setPrompt() {
	m.rl.SetPrompt(term.Blue("yanpassword") + "> ")
}

func (m *Manager) cmd(line string) {
	var args []string
	var argsLine string

	line = strings.Trim(line, " \n\t")

	cmdRunes, rest := wsSplit([]rune(line))
	cmd := string(cmdRunes)

	if cmd == "" {
		return
	}

	if rest == nil {
		args = make([]string, 0)
		argsLine = ""
	} else {
		argsLine = string(rest)
		args = exprWhiteSpace.Split(argsLine, -1)
	}

	if handler, ok := m.handlers[cmd]; ok {
		handler(cmd, argsLine, args...)
	} else {
		term.Errorf("Unknown command: %s\n", cmd)
	}
}
