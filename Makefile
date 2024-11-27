.PHONY: build test clean

# 构建
build:
	go build -o bin/gobase cmd/main.go

# 测试
test:
	go test -v ./...

# 清理
clean:
	rm -rf bin/
	rm -rf dist/

# 运行
run:
	go run cmd/main.go

# 依赖更新
deps:
	go mod tidy
	go mod vendor

# 代码格式化
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run