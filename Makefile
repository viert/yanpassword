MAIN = src/yanpassword.go
DEPS = github.com/chzyer/readline \
		github.com/studio-b12/gowebdav \
		golang.org/x/crypto/pbkdf2

ENV = GOPATH=$(CURDIR)
SOURCE = src/yanpassword.go \
		src/term/colors.go \
		src/term/util.go \
		src/manager/auth_data.go \
		src/manager/handlers.go \
		src/manager/manager.go \
		src/manager/passdb.go \
		src/manager/readline.go \
		src/crypter/crypter.go \
		src/client/client.go

yanpassword: $(SOURCE)
	env $(ENV) go build $(MAIN)

deps:
	env $(ENV) go get $(DEPS)

clean:
	rm -f yanpassword
	find pkg -name *.a -delete
