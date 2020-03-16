executable := cleaner-linux-amd64

build:
	@echo Building $(executable)
	GOOS=linux GO111MODULE=on CGO_ENABLED=0 go build -o $(executable)
	# docker build -t someprefix/$(executable) -f Dockerfile .
	docker build -t cleaner .

run:
	docker run --rm -v /Volumes/RAMDisk:/data cleaner --debug=true --paths=/data