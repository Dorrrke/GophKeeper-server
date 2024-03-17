build: download
	go mod tidy
	go env -w CGO_ENABLED=1
	go env -w GOOS=linux
	go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildCommit=167f7db -X 'main.buildDate=$(date +'%m/%d/%Y')'" -o gophkeeper-server ./cmd/kepeerserver/main.go

build-win: download
	go mod tidy
	go env -w CGO_ENABLED=1
	go build -ldflags "-X main.buildVersion=v1.0.0 -X main.buildCommit=167f7db -X 'main.buildDate=$(date +'%m/%d/%Y')'" -o gophkeeper-server.exe ./cmd/kepeerserver/main.go

download:
	go mod download
	go mod verify

run: download
	go run ./cmd/kepeerserver/main.go --debug

