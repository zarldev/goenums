build:
	go build -o bin/goenums cmd/goenums.go

install:
	chmod +x bin/goenums
	cp bin/goenums /usr/local/go/bin/goenums

