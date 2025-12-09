# Copyright 2022 CloudWeGo Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: all basic advanced clean install test

# 默认目标
all: basic advanced

# 安装 thriftgo
install:
	cd ../.. && make build_tool

# 生成基本测试的 TypeScript 代码
basic: install
	@echo "生成基本测试的 TypeScript 代码..."
	@mkdir -p gen-basic
	@../../thriftgo.exe -g typescript -r -o gen-basic test_basic.thrift
	@echo "基本测试代码生成完成，输出目录: gen-basic/"

merchants: install clean
	@echo "生成商户测试的 TypeScript 代码..."
	@mkdir -p gen-merchants
	@../../thriftgo.exe -g typescript -r -o gen-merchants d:/huidu/xym_idl/idl/http/http_admin/merchant.thrift
	@echo "商户测试代码生成完成，输出目录: gen-merchants/"
merchants1: install clean
	@echo "生成商户测试的 TypeScript 代码..."
	@mkdir -p gen-merchants
	@../../thriftgo.exe -g typescript:snake_style_property_name=true -r -o gen-merchants d:/huidu/xym_idl/idl/http/http_admin/merchant.thrift
	@echo "商户测试代码生成完成，输出目录: gen-merchants/"
merchants2: install clean
	@echo "生成商户测试的 TypeScript 代码..."
	@mkdir -p gen-merchants
	@../../thriftgo.exe -g typescript:lower_camel_case_property_name=true -r -o gen-merchants d:/huidu/xym_idl/idl/http/http_admin/merchant.thrift
	@echo "商户测试代码生成完成，输出目录: gen-merchants/"
enum_test: install clean
	@echo "生成商户测试的 TypeScript 代码..."
	@mkdir -p gen-enum_test
	@../../thriftgo.exe -g typescript:lower_camel_case_property_name=true -r -o gen-enum_test test_enum_tag.thrift
	@echo "商户测试代码生成完成，输出目录: gen-enum_test/"

default_values_test: install clean
	@echo "生成商户测试的 TypeScript 代码..."
	@mkdir -p gen-default_values_test
	@../../thriftgo.exe -g typescript:lower_camel_case_property_name=true -r -o gen-default_values_test test_default_values.thrift
	@echo "商户测试代码生成完成，输出目录: gen-default_values_test/"

req_test: install clean
	@echo "生成商户测试的 TypeScript 代码..."

patch_test: install clean
	@echo "生成 PATCH 测试的 TypeScript 代码..."
	@mkdir -p gen-patch
	@../../thriftgo.exe -g typescript -r -o gen-patch test_patch.thrift
	@echo "PATCH 测试代码生成完成，输出目录: gen-patch/"
	@mkdir -p gen-request
	@../../thriftgo.exe -g typescript:lower_camel_case_property_name=true -r -o gen-request ./test_expreq.thrift
	@echo "商户测试代码生成完成，输出目录: gen-request/"

fields_test: install clean
	@echo "生成 fields.ts 测试的 TypeScript 代码..."
	@mkdir -p gen-fields
	@if [ -f ../../thriftgo.exe ]; then \
		../../thriftgo.exe -g typescript -r -o gen-fields test_fields.thrift; \
	else \
		../../thriftgo -g typescript -r -o gen-fields test_fields.thrift; \
	fi
	@echo "fields.ts 测试代码生成完成，输出目录: gen-fields/"
	@echo "检查生成的文件..."
	@if [ -d gen-fields/test/fields ]; then \
		ls -la gen-fields/test/fields/ || echo "检查生成目录"; \
	else \
		echo "生成目录不存在，检查根目录..."; \
		ls -la gen-fields/ || echo "检查生成目录"; \
	fi
	@echo "检查是否生成了 fields.ts 文件..."
	@if [ -f gen-fields/test/fields/merchantsettinginfo.fields.ts ]; then \
		echo "✓ merchantsettinginfo.fields.ts 已生成"; \
		head -20 gen-fields/test/fields/merchantsettinginfo.fields.ts; \
	else \
		echo "✗ merchantsettinginfo.fields.ts 未生成"; \
	fi
	@if [ -f gen-fields/test/fields/userinfo.fields.ts ]; then \
		echo "✓ userinfo.fields.ts 已生成"; \
	else \
		echo "✗ userinfo.fields.ts 未生成"; \
	fi
	@if [ -f gen-fields/test/fields/orderinfo.fields.ts ]; then \
		echo "✓ orderinfo.fields.ts 已生成"; \
	else \
		echo "✗ orderinfo.fields.ts 未生成"; \
	fi
	@if [ -f gen-fields/test/fields/extendedinfo.fields.ts ]; then \
		echo "✓ extendedinfo.fields.ts 已生成"; \
	else \
		echo "✗ extendedinfo.fields.ts 未生成"; \
	fi
	@if [ ! -f gen-fields/test/fields/productinfo.fields.ts ]; then \
		echo "✓ productinfo.fields.ts 未生成（正确）"; \
	else \
		echo "✗ productinfo.fields.ts 不应该生成"; \
	fi

# 生成高级测试的 TypeScript 代码
advanced: install
	@echo "生成高级测试的 TypeScript 代码..."
	@mkdir -p gen-advanced
	@../../thriftgo.exe -g typescript -r -o gen-advanced test_advanced.thrift
	@echo "高级测试代码生成完成，输出目录: gen-advanced/"

# 生成所有 TypeScript 代码
gen: basic advanced
	@echo "所有 TypeScript 代码生成完成"

# 测试生成的代码
test: gen
	@echo "测试生成的 TypeScript 代码..."
	@echo "检查基本测试文件..."
	@ls -la gen-basic/
	@echo "检查高级测试文件..."
	@ls -la gen-advanced/
	@echo "检查基本测试的展开文件..."
	@ls -la gen-basic/test_basic/ || echo "基本测试未使用展开模式"
	@echo "检查高级测试的展开文件..."
	@ls -la gen-advanced/test_advanced/ || echo "高级测试未使用展开模式"

# 清理生成的文件
clean:
	@echo "清理生成的 TypeScript 文件..."
	@rm -rf gen-* 
	@echo "清理完成"

# 显示帮助信息
help:
	@echo "TypeScript 代码生成 Makefile"
	@echo ""
	@echo "可用目标:"
	@echo "  all        - 生成所有 TypeScript 代码 (默认)"
	@echo "  basic      - 生成基本测试的 TypeScript 代码"
	@echo "  advanced   - 生成高级测试的 TypeScript 代码"
	@echo "  fields_test - 生成 fields.ts 测试的 TypeScript 代码"
	@echo "  gen        - 生成所有 TypeScript 代码 (同 all)"
	@echo "  test       - 测试生成的代码"
	@echo "  clean      - 清理生成的文件"
	@echo "  install    - 安装 thriftgo"
	@echo "  help       - 显示此帮助信息"
	@echo ""
	@echo "示例用法:"
	@echo "  make basic     # 只生成基本测试"
	@echo "  make advanced  # 只生成高级测试"
	@echo "  make test      # 生成并测试所有代码"
	@echo "  make clean     # 清理所有生成的文件"