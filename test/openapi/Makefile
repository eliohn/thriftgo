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
# 安装 thriftgo
install:
	cd ../.. && make build_tool
# 清理生成的文件
clean:
	@echo "清理生成的 TypeScript 文件..."
	@rm -rf gen-* 
	@echo "清理完成"

merchants: install clean
	@echo "生成商户测试的 openapi 代码..."
	@mkdir -p gen-merchants
	@../../thriftgo.exe -g openapi -r -o gen-merchants d:/huidu/xym_idl/idl/http/http_admin/merchant.thrift
	@echo "商户测试代码生成完成，输出目录: gen-merchants/"