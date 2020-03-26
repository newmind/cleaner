executable := scanner-linux-amd64

build:
	@echo Building $(executable)
	GOOS=linux GO111MODULE=on CGO_ENABLED=0 go build -o $(executable)

run:
	$(executable) --path=/data