binary:
	GOOS=windows GOARCH=amd64 go build -o bin/organizer_windows_amd64 organizer.go
	GOOS=darwin GOARCH=amd64 go build -o bin/organizer_mac_amd64 organizer.go
	GOOS=linux GOARCH=amd64 go build -o bin/organizer_linux_amd64 organizer.go
