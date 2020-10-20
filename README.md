## yanpassword

### DEPRECATION WARNING
Yanpassword is going to break any moment as the Yandex.Disk WebDAV API is deprecated for a while.
Consider moving to https://github.com/viert/kstore

### Description

Yanpassword is a command-line tool for storing services sensitive data in a safe way backed up by Yandex.Disk storage.

### Build

You will need a golang compiler. The building process is as easy as typing this command below:

```
go build -o yanpassword cmd/yanpassword/main.go
```

### How it works

First time you run the program it will ask you to create a master password. Then you have to type in your yandex account credentials, it's highly recommended though to use a separate password for the application which you can set up at https://passport.yandex.ru/profile in "Application passwords" section. If you choose to do that be sure to create a new application password with Yandex.Disk/Webdav access only.

After you type in your yandex credentials, the program will check their validity and store the authentication data encrypted with your master password in `~/.yanpasswd_auth` file. Next time you use the application you won't need yandex credentials, only the master password.

Your service auth data is stored in your Yandex.Disk in `.yanpassword` folder. Files with data are named `db.bin`, `db.bin.1`, `db.bin.2` and so on up to 5. The actual data is stored in `db.bin`, the others are backup files. Those files are encrypted JSON-files of the given structure:

```
{
    "<serviceName1>": {
        "name": "<serviceName1>",
        "username": ...,
        "password": ...,
        "comment": ...,
        "updated_at": ...,
        "url": ...
    },
    ...
}
```

### Commands

`ls`, `list` list all the service names you have

`get <servicename>` prints all the data you entered previously about the service

`getpass <servicename>` prints only the password of the given service

`set <servicename>` is a command to modify the given service or create a new one while `setpass <servicename>` will only change password of an _existing_ service.

`del`, `delete`, `remove`, `rm` are aliases to remove a service from the list

All changes become persistent only after typing the `save` command. Save will encrypt data with your master password you typed on start, move the backup files and save the actual data in `db.bin`

### Change Master Password

Yanpassword is supposed to have a `chpass` command in future versions. The way to change the master password now is following:

- remove `~/.yanpasswd_auth` file so the app thinks you start it for the first time
- run the app and create a new master password
- enter yandex credentials once again
- yanpassword will connect to Yandex.Disk and load the data but won't be able to decrypt it as the master password has changed
- yanpassword will ask for a previous master password to decrypt data
- check that your data is loaded and valid
- type `save` to save the data encrypted with your new master password

Remember that backup files remain encrypted with the previous version of MP.

### Encryption

Data is encrypted with AES key based on a pbkdf2 passkey with a MD5-based salt.

### Awaited features

- `chpass` command to change the master password from inside the app in a convenient way
- automatic `updated_at` field to keep track of changes (at the moment the field is not used)

### Migrating

If you were using the python version of yanpassword, the way to migrate is the following:

- checkout the `dataexport` branch of https://github.com/viert/yanpassword-legacy repo
- run the `yanpassword.py` as usual
- run the `export` command (available only in `dataexport` branch)
- save the output json to a file
- run the brand new yanpassword
- use the `import <file.json>` command to import the data
- ensure the data is imported properly and type `save` command to save the result
