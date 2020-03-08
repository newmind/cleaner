executable := cleaner-linux-amd64

build:
	@echo Building $(executable)
	GOOS=linux GO111MODULE=on CGO_ENABLED=0 go build -o $(executable)
	# docker build -t someprefix/$(executable) -f Dockerfile .
