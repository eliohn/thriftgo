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

package typescript

// 参数定义
type param struct {
	name string
	desc string
}

var allParams = []param{
	{
		name: "skip_empty",
		desc: "跳过空文件生成",
	},
	{
		name: "generate_interfaces",
		desc: "生成接口定义",
	},
	{
		name: "generate_classes",
		desc: "生成类定义",
	},
	{
		name: "use_strict_mode",
		desc: "使用严格模式",
	},
	{
		name: "use_es6_modules",
		desc: "使用 ES6 模块",
	},
	{
		name: "output_dir",
		desc: "输出目录",
	},
}
