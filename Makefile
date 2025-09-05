.PHONY: build run test clean setup

# Go 参数
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=insider-monitor
BINARY_UNIX=$(BINARY_NAME)_unix

# 构建目录
BUILD_DIR=bin

all: test build

build: 
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) -v ./cmd/monitor

clean: 
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf data/*.log

run:
	$(GOCMD) run cmd/monitor/main.go

test: 
	$(GOTEST) -v ./...

setup:
	$(GOGET) -v ./...
	cp -n config.example.json config.json || true

# 交叉编译
build-linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_UNIX) -v ./cmd/monitor

docker-build:
	docker build -t $(BINARY_NAME) .

# 帮助目标
help:
        @echo "可用目标:"
        @echo "  make build       - 构建二进制文件"
        @echo "  make run        - 运行应用程序"
        @echo "  make test       - 运行测试"
        @echo "  make clean      - 清理构建文件"
        @echo "  make setup      - 初始设置"
        @echo "  make build-linux- 为 Linux 构建"
