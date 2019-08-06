package client

import (
	"crypter"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/studio-b12/gowebdav"
)

const (
	webdavURL  = "https://webdav.yandex.ru"
	passdbDir  = ".yanpassword"
	passdbFile = "db.bin"

	maxBackups = 5
)

// CheckAuth checks auth with a dummy yandex webdav request
func CheckAuth(authData *crypter.AuthData) error {
	cli := gowebdav.NewClient(webdavURL, authData.Username, authData.Password)
	_, err := cli.ReadDir("/")
	if err != nil {
		return err
	}
	return nil
}

// PassdbClient syncs passdb data with yandex disk as well as creates backups and stuff
type PassdbClient struct {
	cli *gowebdav.Client
}

// NewPassdbClient creates a new instance of PassdbClient
func NewPassdbClient(authData *crypter.AuthData) *PassdbClient {
	pdbc := new(PassdbClient)
	pdbc.cli = gowebdav.NewClient(webdavURL, authData.Username, authData.Password)
	return pdbc
}

// Load loads the main passdb file
func (pdbc *PassdbClient) Load() ([]byte, error) {
	filename := path.Join(passdbDir, passdbFile)
	data, err := pdbc.cli.Read(filename)
	return data, err
}

// Is404 tries to figure out if the error is a 404 not found error
func Is404(err error) bool {
	if err == nil {
		return false
	}

	pe := err.(*os.PathError)
	if pe == nil {
		return false
	}

	return strings.HasPrefix(pe.Err.Error(), "404")
}

// Save saves the main passddb file contents, making backups of previous files
func (pdbc *PassdbClient) Save(data []byte) error {
	var prev string
	var next string

	filename := path.Join(passdbDir, passdbFile)
	fmt.Printf("Creating backups")
	for i := maxBackups - 1; i > 0; i-- {
		prev = fmt.Sprintf("%s.%d", filename, i)
		next = fmt.Sprintf("%s.%d", filename, i+1)

		fmt.Printf(".")
		_, err := pdbc.cli.Stat(prev)
		if Is404(err) {
			continue
		}

		err = pdbc.cli.Rename(prev, next, true)
		if err != nil {
			fmt.Printf("\nError moving backup %s to %s: %s\n", prev, next, err)
			return err
		}
	}

	fmt.Printf(".")
	prev = filename
	next = fmt.Sprintf("%s.1", filename)
	if _, err := pdbc.cli.Stat(prev); err == nil {
		err = pdbc.cli.Rename(prev, next, true)
		if err != nil {
			fmt.Printf("\nError backing up file %s to %s: %s\n", prev, next, err)
			return err
		}
	}
	fmt.Println("\nSaving data...")
	return pdbc.cli.Write(filename, data, os.FileMode(0644))
}
