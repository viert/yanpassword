package manager

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"term"
)

func (m *Manager) setupHandlers() {
	m.handlers = make(map[string]cmdHandler)
	m.handlers["exit"] = m.doExit
	m.handlers["save"] = m.doSave
	m.handlers["import"] = m.doImport
	m.handlers["list"] = m.doList
	m.handlers["ls"] = m.doList
	m.handlers["get"] = m.doGet
	m.handlers["getpass"] = m.doGet
	m.handlers["set"] = m.doSet
	m.handlers["setpass"] = m.doSet
}

func (m *Manager) doExit(name string, argsLine string, args ...string) {
	m.stopped = true
}

func (m *Manager) doSave(name string, argsLine string, args ...string) {
	m.savePassdb()
}

func (m *Manager) doImport(name string, argsLine string, args ...string) {
	if len(args) < 1 {
		term.Errorf("import command requires a file name to import from\n")
		return
	}

	filename := args[0]

	f, err := os.Open(filename)
	if err != nil {
		term.Errorf("error opening file %s: %s\n", filename, err)
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		term.Errorf("error reading file %s: %s\n", filename, err)
		return
	}

	sd := make(serviceData)
	err = json.Unmarshal(data, &sd)
	if err != nil {
		term.Errorf("error parsing file %s: %s\n", filename, err)
		return
	}

	added := 0
	skipped := 0
	for k, v := range sd {
		if _, found := m.data[k]; found {
			term.Warnf("Service %s already exists, skipping\n", k)
			skipped++
			continue
		}
		m.data[k] = v
		added++
	}

	term.Successf("%d items imported, %d items skipped\n", added, skipped)
	term.Warnf("The imported data is not persistent yet, don't forget to **save** it.\n")
}

func (m *Manager) doList(name string, argsLine string, args ...string) {
	if len(m.data) > 0 {
		names := make([]string, len(m.data))
		i := 0
		for k := range m.data {
			names[i] = k
			i++
		}
		sort.Strings(names)

		for _, name := range names {
			fmt.Println(name)
		}
	}
}

func (m *Manager) doGet(name string, argsLine string, args ...string) {
	if len(args) < 1 {
		term.Errorf("%s command requires a service name\n", name)
		return
	}

	serviceName := args[0]
	if item, found := m.data[serviceName]; found {
		switch name {
		case "getpass":
			fmt.Println(item.Password)
		default:
			fmt.Printf("Service: %s\n", item.Name)
			if item.Username != "" {
				fmt.Printf("Username: %s\n", item.Username)
			}
			if item.Password != "" {
				fmt.Printf("Password: %s\n", item.Password)
			}
			if item.Comment != "" {
				fmt.Printf("Comment: %s\n", item.Comment)
			}
			if item.URL != "" {
				fmt.Printf("URL: %s\n", item.URL)
			}
			fmt.Println()
		}
	} else {
		term.Errorf("Service %s not found\n", serviceName)
	}
}

func (m *Manager) doSet(name string, argsLine string, args ...string) {
}
