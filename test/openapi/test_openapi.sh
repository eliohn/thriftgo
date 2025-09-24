#!/bin/bash

# OpenAPI 生成器测试脚本

echo "测试 OpenAPI 生成器..."

# 创建输出目录
mkdir -p output

# 测试基本生成
echo "1. 测试基本生成..."
go run ../../main.go -g openapi -o output test_basic.thrift
if [ $? -eq 0 ]; then
    echo "✓ 基本生成成功"
    echo "生成的文档："
    head -20 output/test_basic.yaml
else
    echo "✗ 基本生成失败"
fi

echo ""

# 测试高级生成
echo "2. 测试高级生成..."
go run ../../main.go -g openapi -o output test_advanced.thrift
if [ $? -eq 0 ]; then
    echo "✓ 高级生成成功"
    echo "生成的文档行数："
    wc -l output/test_advanced.yaml
else
    echo "✗ 高级生成失败"
fi

echo ""

# 测试带配置选项的生成
echo "3. 测试带配置选项的生成..."
go run ../../main.go -g openapi -o output -p title=UserServiceAPI -p base_path=/api/v1 test_basic.thrift
if [ $? -eq 0 ]; then
    echo "✓ 带配置选项的生成成功"
    echo "生成的文档标题："
    grep "title:" output/test_basic.yaml
else
    echo "✗ 带配置选项的生成失败"
fi

echo ""
echo "测试完成！"
