# OpenAPI 生成器测试 Makefile

.PHONY: test clean

# 测试 OpenAPI 生成器
test:
	@echo "测试 OpenAPI 生成器..."
	@mkdir -p output
	@cd .. && go run main.go -g openapi -o test/openapi/output test/openapi/test_basic.thrift
	@echo "生成的 OpenAPI 文档："
	@cat output/test_basic.yaml

# 清理生成的文件
clean:
	@rm -rf output

# 显示帮助
help:
	@echo "可用的命令："
	@echo "  test  - 测试 OpenAPI 生成器"
	@echo "  clean - 清理生成的文件"
	@echo "  help  - 显示此帮助信息"
