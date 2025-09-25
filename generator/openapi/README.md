# OpenAPI 生成器

这个生成器可以将 Thrift IDL 文件转换为 OpenAPI 3.0 规范的文档。

## 功能特性

- 支持将 Thrift 结构体转换为 OpenAPI Schema
- 支持将 Thrift 服务转换为 OpenAPI Paths
- 支持枚举类型转换
- 支持基本类型映射
- 支持容器类型（list、map、set）
- 支持自定义配置选项

## 使用方法

### 基本用法

```bash
thriftgo -g openapi -o output_dir your_file.thrift
```

### 配置选项

生成器支持以下配置选项：

- `skip_empty`: 跳过生成空文件 (默认: false)
- `version`: OpenAPI 规范版本 (默认: 3.0.0)
- `title`: API 标题 (默认: Thrift API)
- `base_path`: API 基础路径 (默认: /)
- `description`: API 描述
- `contact_name`: 联系人姓名
- `contact_email`: 联系人邮箱
- `contact_url`: 联系人网址
- `license_name`: 许可证名称
- `license_url`: 许可证网址
- `server_url`: 服务器 URL
- `server_description`: 服务器描述

### 示例

```bash
# 基本生成
thriftgo -g openapi -o docs user_service.thrift

# 带配置选项的生成
thriftgo -g openapi -o docs \
  -p title="用户服务 API" \
  -p description="用户管理相关的 API 接口" \
  -p base_path="/api/v1" \
  -p contact_name="开发团队" \
  -p contact_email="dev@example.com" \
  user_service.thrift
```

## 类型映射

### 基本类型映射

| Thrift 类型 | OpenAPI 类型 | 格式 |
|------------|-------------|------|
| bool | boolean | - |
| byte | integer | int8 |
| i16 | integer | int16 |
| i32 | integer | int32 |
| i64 | integer | int64 |
| double | number | double |
| string | string | - |
| binary | string | binary |

### 容器类型映射

| Thrift 类型 | OpenAPI 类型 |
|------------|-------------|
| list<T> | array |
| map<K,V> | object |
| set<T> | array (uniqueItems: true) |

### 复杂类型映射

| Thrift 类型 | OpenAPI 类型 |
|------------|-------------|
| struct | object |
| union | object |
| exception | object |
| enum | string (with enum values) |

## 服务方法映射

生成器会根据方法名称自动推断 HTTP 方法：

- 以 `get`、`find`、`list` 开头的方法 → GET
- 以 `create`、`add`、`insert` 开头的方法 → POST
- 以 `update`、`modify` 开头的方法 → PUT
- 以 `delete`、`remove` 开头的方法 → DELETE
- 其他方法 → POST

## 路径生成

API 路径的生成规则：

1. 基础路径：使用 `base_path` 配置选项
2. 服务路径：服务名转换为 kebab-case
3. 方法路径：方法名转换为 kebab-case

例如：
- 服务名：`UserService`
- 方法名：`getUser`
- 生成路径：`/api/user-service/get-user`

## 示例输出

生成的 OpenAPI 文档包含：

1. **Info 部分**：API 基本信息
2. **Paths 部分**：所有服务方法的 API 路径
3. **Components/Schemas 部分**：所有数据结构的定义

每个 API 路径包含：
- HTTP 方法
- 参数定义
- 请求体定义（如果有）
- 响应定义
- 错误响应定义

## 注意事项

1. 生成器会为每个 Thrift 文件生成一个对应的 YAML 文件
2. 如果 Thrift 文件中没有服务定义，只会生成 Schema 部分
3. 如果 Thrift 文件中没有结构体定义，只会生成 Paths 部分
4. 生成的文档符合 OpenAPI 3.0 规范，可以在 Swagger UI 等工具中查看
