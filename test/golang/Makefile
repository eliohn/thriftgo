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

.PHONY: all unknown cases clean

all: unknown cases

unknown:
	cd unknown_fields && ./run_test.sh

cases:
	cd cases_and_options && ./run_test.sh

clean2:
	@find . -name "gen-*" -type d | while read d; do echo rm -r $$d; rm -r $$d; done

# 清理生成的文件
clean:
	@echo "清理生成的 gen 文件..."
	@rm -rf gen-*
	@echo "清理完成"

# 安装 thriftgo
install:
	cd ../.. && make build_tool

# 生成基本测试的 Golang 代码
basic1: install clean
	@echo "生成基本测试的 Golang 代码..."
	@mkdir -p gen-basic
	@../../thriftgo.exe -g go -thrift frugal_tag,template=slim -r -o gen-basic test_basic.thrift
	@echo "基本测试代码生成完成，输出目录: gen-basic/"

expand1: install clean
	@echo "生成扩展测试的 Golang 代码..."
	@../../thriftgo.exe -g go:lower_camel_style_json_tag -r -o gen-expand test_expand.thrift
	@echo "扩展测试代码生成完成，输出目录: gen-expand/"

expand2: install clean
	@echo "生成扩展测试的 Golang 代码..."
	@../../thriftgo.exe -g go:snake_style_json_tag -r -o gen-expand test_expand.thrift
	@echo "扩展测试代码生成完成，输出目录: gen-expand/"

expreq: install clean
	@echo "生成扩展测试的 Golang 代码..."
	@../../thriftgo.exe -g go:snake_style_json_tag -r -o gen-expreq test_expreq.thrift
	@echo "扩展测试代码生成完成，输出目录: gen-expreq/"