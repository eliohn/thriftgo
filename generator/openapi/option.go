// Copyright 2024 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package openapi

// Parameter represents a generator parameter.
type Parameter struct {
	name string
	desc string
}

var allParams = []Parameter{
	{
		name: "skip_empty",
		desc: "跳过生成空文件",
	},
	{
		name: "version",
		desc: "OpenAPI 规范版本 (默认: 3.0.0)",
	},
	{
		name: "title",
		desc: "API 标题 (默认: Thrift API)",
	},
	{
		name: "base_path",
		desc: "API 基础路径 (默认: /api)",
	},
	{
		name: "description",
		desc: "API 描述",
	},
	{
		name: "contact_name",
		desc: "联系人姓名",
	},
	{
		name: "contact_email",
		desc: "联系人邮箱",
	},
	{
		name: "contact_url",
		desc: "联系人网址",
	},
	{
		name: "license_name",
		desc: "许可证名称",
	},
	{
		name: "license_url",
		desc: "许可证网址",
	},
	{
		name: "server_url",
		desc: "服务器 URL",
	},
	{
		name: "server_description",
		desc: "服务器描述",
	},
	{
		name: "snake_style_property_name",
		desc: "使用 snake_case 命名属性",
	},
	{
		name: "lower_camel_case_property_name",
		desc: "使用 lowerCamelCase 命名属性（默认）",
	},
}
