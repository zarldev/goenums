build:
	go build -o bin/goenums goenums.go

install:
	chmod +x bin/goenums
	cp bin/goenums /usr/local/go/bin/goenums
	
test:
	go test -v ./...

