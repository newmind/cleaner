executable := generator-linux-amd64

build:
	@echo Building $(executable)
	GOOS=linux GO111MODULE=on CGO_ENABLED=0 go build -o $(executable)

run:
	$(executable) -size=1024 -interval=100ms -duration=300 -users=50 -path=/data