test:
	cd ../ && docker build --platform="linux/amd64" -f go-libs/Dockerfile .

build:
	go mod tidy && go build ./...