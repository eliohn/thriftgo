// Copyright 2021 CloudWeGo Authors
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

package golang

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	thrift_option "github.com/cloudwego/thriftgo/extension/thrift_option"

	"github.com/cloudwego/thriftgo/generator/golang/common"
	"github.com/cloudwego/thriftgo/generator/golang/streaming"
	"github.com/cloudwego/thriftgo/parser"
	"github.com/cloudwego/thriftgo/pkg/namespace"
)

const (
	// A prefix to denote synthesized identifiers.
	prefix = "$"
	// nestedAnnotation is to denote the field is nested type.
	nestedAnnotation    = "thrift.nested"
	interfaceAnnotation = "thrift.is_interface"
	aliasAnnotation     = "thrift.is_alias"
	// expandAnnotation is to denote the field should be expanded into parent struct.
	expandAnnotation = "thrift.expand"
)

func _p(id string) string {
	return prefix + id
}

// newScope creates an uninitialized scope from the given IDL.
func newScope(ast *parser.Thrift) *Scope {
	return &Scope{
		ast:       ast,
		imports:   newImportManager(),
		globals:   namespace.NewNamespace(namespace.UnderscoreSuffix),
		namespace: ast.GetNamespaceOrReferenceName("go"),
	}
}

func (s *Scope) init(cu *CodeUtils) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("err = %v, stack = %s", r, debug.Stack())
		}
	}()

	if cu.Features().UseOption {
		for ast := range s.AST().DepthFirstSearch() {
			er := thrift_option.CheckOptionGrammar(ast)
			if er != nil {
				return er
			}
		}
	}

	if cu.Features().ReorderFields {
		for _, x := range s.ast.GetStructLikes() {
			diff := reorderFields(x)
			if diff != nil && diff.original != diff.arranged {
				cu.Info(fmt.Sprintf("<reorder>(%s) %s: %d -> %d: %.2f%%",
					s.ast.Filename, x.Name, diff.original, diff.arranged, diff.percent()))
			}
		}
	}
	s.imports.init(cu, s.ast)
	s.buildIncludes(cu)
	if err = s.installNames(cu); err != nil {
		return err
	}
	s.resolveTypesAndValues(cu)
	return nil
}

func (s *Scope) buildIncludes(cu *CodeUtils) {
	// the indices of includes must be kept because parser.Reference.Index counts the unused IDLs.
	cnt := len(s.ast.Includes)
	s.includes = make([]*Include, cnt)

	for idx, inc := range s.ast.Includes {
		if !inc.GetUsed() {
			continue
		}
		s.includes[idx] = s.include(cu, inc.Reference)
	}
}

func (s *Scope) include(cu *CodeUtils, t *parser.Thrift) *Include {
	scope, err := BuildScope(cu, t)
	if err != nil {
		panic(err)
	}
	pth := scope.importPath
	pkg := scope.importPackage
	if s.namespace != scope.namespace {
		pkg = s.imports.Add(pkg, pth)
	}
	return &Include{
		PackageName: pkg,
		ImportPath:  pth,
		Scope:       scope,
	}
}

// includeIDL adds an probably new IDL to the include list.
func (s *Scope) includeIDL(cu *CodeUtils, t *parser.Thrift) (pkgName string) {
	_, pth := cu.Import(t)
	if pkgName = s.imports.Get(pth); pkgName != "" {
		return
	}
	inc := s.include(cu, t)
	s.includes = append(s.includes, inc)
	return inc.PackageName
}

func (s *Scope) installNames(cu *CodeUtils) error {
	for _, v := range s.ast.Services {
		if err := s.buildService(cu, v); err != nil {
			return err
		}
	}
	for _, v := range s.ast.GetStructLikes() {
		s.buildStructLike(cu, v)
	}
	for _, v := range s.ast.Enums {
		s.buildEnum(cu, v)
	}
	for _, v := range s.ast.Typedefs {
		s.buildTypedef(cu, v)
	}
	for _, v := range s.ast.Constants {
		s.buildConstant(cu, v)
	}
	return nil
}

func (s *Scope) identify(cu *CodeUtils, raw string) string {
	name, err := cu.Identify(raw)
	if err != nil {
		panic(err)
	}
	if !strings.HasPrefix(raw, prefix) && cu.Features().CompatibleNames {
		if strings.HasPrefix(name, "New") || strings.HasSuffix(name, "Args") || strings.HasSuffix(name, "Result") {
			name += "_"
		}
	}
	return name
}

func (s *Scope) buildService(cu *CodeUtils, v *parser.Service) error {
	// service name
	sn := s.identify(cu, v.Name)
	sn = s.globals.Add(sn, v.Name)

	svc := &Service{
		Service: v,
		scope:   namespace.NewNamespace(namespace.UnderscoreSuffix),
		from:    s,
		name:    Name(sn),
	}
	s.services = append(s.services, svc)

	// function names
	for _, f := range v.Functions {
		fn := s.identify(cu, f.Name)
		fn = svc.scope.Add(fn, f.Name)
		st, err := streaming.ParseStreaming(f)
		if err != nil {
			return fmt.Errorf("service %s: %s", v.Name, err.Error())
		}

		svc.functions = append(svc.functions, &Function{
			Function:  f,
			scope:     namespace.NewNamespace(namespace.UnderscoreSuffix),
			name:      Name(fn),
			service:   svc,
			streaming: st,
		})
	}

	// install names for argument types and response types
	for idx, f := range v.Functions {
		argType, resType := buildSynthesized(f)
		an := v.Name + s.identify(cu, _p(f.Name+"_args"))
		rn := v.Name + s.identify(cu, _p(f.Name+"_result"))

		fun := svc.functions[idx]
		fun.argType = s.buildStructLike(cu, argType, _p(an))
		if !f.Oneway {
			fun.resType = s.buildStructLike(cu, resType, _p(rn))
			if !f.Void {
				fun.resType.fields[0].isResponse = true
			}
		}

		s.buildFunction(cu, fun, f)
	}

	// install names for client and processor
	cn := sn + "Client"
	pn := sn + "Processor"
	s.globals.MustReserve(cn, _p("client:"+v.Name))
	s.globals.MustReserve(pn, _p("processor:"+v.Name))
	return nil
}

// buildFunction builds a namespace for parameters of a Function.
// This function is used to resolve conflicts between parameter, receiver and local variables in generated method.
// Template 'Service' and 'FunctionSignature' depend on this function.
func (s *Scope) buildFunction(cu *CodeUtils, f *Function, v *parser.Function) {
	ns := f.scope

	ns.MustReserve("p", _p("p"))     // the receiver of method
	ns.MustReserve("err", _p("err")) // error
	ns.MustReserve("ctx", _p("ctx")) // first parameter

	if !v.Void {
		ns.MustReserve("r", _p("r"))             // response
		ns.MustReserve("_result", _p("_result")) // a local variable
	}

	for _, a := range v.Arguments {
		name := common.LowerFirstRune(s.identify(cu, a.Name))
		if isKeywords[name] {
			name = "_" + name
		}
		ns.Add(name, a.Name)
	}

	for _, t := range v.Throws {
		name := common.LowerFirstRune(s.identify(cu, t.Name))
		if isKeywords[name] {
			name = "_" + name
		}
		ns.Add(name, t.Name)
	}
}

func (s *Scope) buildTypedef(cu *CodeUtils, t *parser.Typedef) {
	tn := s.identify(cu, t.Alias)
	tn = s.globals.Add(tn, t.Alias)
	if t.Type.Category.IsStructLike() {
		fn := "New" + tn
		s.globals.MustReserve(fn, _p("new:"+t.Alias))
	}
	s.typedefs = append(s.typedefs, &Typedef{
		Typedef: t,
		name:    Name(tn),
	})
}

func (s *Scope) buildEnum(cu *CodeUtils, e *parser.Enum) {
	en := s.identify(cu, e.Name)
	en = s.globals.Add(en, e.Name)

	enum := &Enum{
		Enum:  e,
		scope: namespace.NewNamespace(namespace.UnderscoreSuffix),
		name:  Name(en),
	}
	for _, v := range e.Values {
		vn := enum.scope.Add(en+"_"+v.Name, v.Name)
		ev := &EnumValue{
			EnumValue: v,
			name:      Name(vn),
		}
		if cu.Features().TypedEnumString {
			ev.literal = Code(ev.name)
		} else {
			ev.literal = Code(v.Name)
		}
		enum.values = append(enum.values, ev)
	}
	s.enums = append(s.enums, enum)
}

func (s *Scope) buildConstant(cu *CodeUtils, v *parser.Constant) {
	cn := s.identify(cu, v.Name)
	cn = s.globals.Add(cn, v.Name)
	s.constants = append(s.constants, &Constant{
		Constant: v,
		name:     Name(cn),
	})
}

func (s *Scope) setRefImport(refPath string) {
	if s == nil {
		return
	}
	s.refPath = refPath
	arr := strings.Split(refPath, "/")
	s.refPackage = arr[len(arr)-1]
}

// getReferencedStruct returns the referenced struct for a field.
func (s *Scope) getReferencedStruct(f *parser.Field) *parser.StructLike {
	var referencedStruct *parser.StructLike

	// Extract package name and type name from f.Type.Name (e.g., "base.MyData" -> "base", "MyData")
	var expectedPackageName, expectedTypeName string
	if strings.Contains(f.Type.Name, ".") {
		parts := strings.Split(f.Type.Name, ".")
		expectedPackageName = strings.Join(parts[:len(parts)-1], ".") // Support nested packages
		expectedTypeName = parts[len(parts)-1]
	} else {
		expectedTypeName = f.Type.Name
	}

	// First try to find by reference (for cross-file references)
	if f.Type.Reference != nil && f.Type.Reference.Index >= 0 {
		// Use the reference's package name if available
		referencePackageName := ""
		if f.Type.Reference.Name != "" {
			referencePackageName = f.Type.Reference.Name
		}

		// For cross-file references, we need to search in includes first
		// because the index refers to the position in the included file
		for _, inc := range s.includes {
			if inc == nil || inc.Scope == nil {
				continue
			}

			// Check if this include matches the expected package name
			if expectedPackageName != "" || referencePackageName != "" {
				// Determine which package name to use for comparison
				packageNameToMatch := referencePackageName
				if packageNameToMatch == "" {
					packageNameToMatch = expectedPackageName
				}
			}

			// Check if the index matches in this include
			if int(f.Type.Reference.Index) < len(inc.Scope.ast.Structs) {
				var ix int
				for index, candidateStruct := range inc.Scope.ast.Structs {
					if candidateStruct.Name == expectedTypeName {
						referencedStruct = candidateStruct
						ix = index
						break
					}
				}
				candidateStruct := inc.Scope.ast.Structs[ix]
				// 验证结构体名称是否匹配字段类型名称
				if candidateStruct.Name == expectedTypeName {
					referencedStruct = candidateStruct
					break
				}
			}
		}

		// If not found in includes, check current file as fallback
		if referencedStruct == nil && int(f.Type.Reference.Index) < len(s.ast.Structs) {
			candidateStruct := s.ast.Structs[f.Type.Reference.Index]
			if candidateStruct.Name == expectedTypeName {
				referencedStruct = candidateStruct
			}
		}
	} else {
		// Find by name in current file
		for _, st := range s.ast.Structs {
			if st.Name == f.Type.Name {
				referencedStruct = st
				break
			}
		}

		// Also check unions and exceptions
		if referencedStruct == nil {
			for _, st := range s.ast.Unions {
				if st.Name == f.Type.Name {
					referencedStruct = st
					break
				}
			}
		}
		if referencedStruct == nil {
			for _, st := range s.ast.Exceptions {
				if st.Name == f.Type.Name {
					referencedStruct = st
					break
				}
			}
		}

		// If not found in current file, search in included files
		if referencedStruct == nil {
			for _, inc := range s.includes {
				if inc == nil || inc.Scope == nil {
					continue
				}

				// Check if package name matches for name-based search
				if expectedPackageName != "" && inc.PackageName != expectedPackageName {
					continue
				}

				// Search in structs
				for _, st := range inc.Scope.ast.Structs {
					if st.Name == expectedTypeName {
						referencedStruct = st
						break
					}
				}
				if referencedStruct != nil {
					break
				}

				// Search in unions
				for _, st := range inc.Scope.ast.Unions {
					if st.Name == expectedTypeName {
						referencedStruct = st
						break
					}
				}
				if referencedStruct != nil {
					break
				}

				// Search in exceptions
				for _, st := range inc.Scope.ast.Exceptions {
					if st.Name == expectedTypeName {
						referencedStruct = st
						break
					}
				}
				if referencedStruct != nil {
					break
				}
			}
		}
	}
	return referencedStruct
}

func (s *Scope) buildStructLike(cu *CodeUtils, v *parser.StructLike, usedName ...string) *StructLike {
	nn := v.Name
	if len(usedName) != 0 {
		nn = usedName[0]
	}
	sn := s.identify(cu, nn)
	sn = s.globals.Add(sn, v.Name)
	s.globals.MustReserve("New"+sn, _p("new:"+nn))

	fids := "fieldIDToName_" + sn
	s.globals.MustReserve(fids, _p("ids:"+nn))

	// built-in methods
	funcs := []string{"Read", "Write", "String"}
	if !strings.HasPrefix(v.Name, prefix) {
		if v.Category == "union" {
			funcs = append(funcs, "CountSetFields")
		}
		if v.Category == "exception" {
			funcs = append(funcs, "Error")
		}
		if cu.Features().KeepUnknownFields {
			funcs = append(funcs, "CarryingUnknownFields")
		}
		if cu.Features().GenDeepEqual {
			funcs = append(funcs, "DeepEqual")
		}
	}

	st := &StructLike{
		StructLike: v,
		scope:      namespace.NewNamespace(namespace.UnderscoreSuffix),
		name:       Name(sn),
	}

	for _, fn := range funcs {
		st.scope.MustReserve(fn, _p(fn))
	}

	id2str := func(id int32) string {
		i := int(id)
		if i < 0 {
			return "_" + strconv.Itoa(-i)
		}
		return strconv.Itoa(i)
	}
	// reserve method names
	for _, f := range v.Fields {
		fn := s.identify(cu, f.Name)
		if cu.Features().EnableNestedStruct && isNestedField(f) {
			// EnableNestedStruct, the type name needs to be used when retrieving the value for getter&setter
			fn = s.identify(cu, f.Type.Name)
			if strings.Contains(fn, ".") {
				fns := strings.Split(fn, ".")
				fn = s.identify(cu, fns[len(fns)-1])
			}
		}

		st.scope.Add("Get"+fn, _p("get:"+f.Name))
		if cu.Features().GenerateSetter {
			st.scope.Add("Set"+fn, _p("set:"+f.Name))
		}
		if SupportIsSet(f) {
			st.scope.Add("IsSet"+fn, _p("isset:"+f.Name))
		}
		id := id2str(f.ID)
		st.scope.Add("ReadField"+id, _p("read:"+id))
		st.scope.Add("writeField"+id, _p("write:"+id))
		if cu.Features().GenDeepEqual {
			st.scope.Add("Field"+id+"DeepEqual", _p("deepequal:"+id))
		}
	}

	// field names
	for _, f := range v.Fields {
		fn := s.identify(cu, f.Name)
		isNested := false
		if cu.Features().EnableNestedStruct && isNestedField(f) {
			isNested = true
		}
		fn = st.scope.Add(fn, f.Name)
		id := id2str(f.ID)
		// Check if this field should be expanded
		isExpandable := false
		var expandedFields []*Field
		if f.Type.Category.IsStructLike() {
			// Check if field has explicit expand annotation OR if the referenced struct is expandable
			shouldExpand := isExpandField(f)

			// Find the referenced struct
			referencedStruct := s.getReferencedStruct(f)
			// If struct is found and either field has explicit expand annotation OR struct is expandable
			if referencedStruct != nil {
				// Check if struct is expandable (has expandable = "true" annotation)
				structIsExpandable := referencedStruct.Expandable != nil && *referencedStruct.Expandable
				//log.Printf("struct %s is expandable: %t , shouldExpand: %t", referencedStruct.Name, structIsExpandable, shouldExpand)
				if shouldExpand || structIsExpandable {
					isExpandable = true
					//fieldNameSpace := ""
					// 如果f.Type.Name 有带命名空间，把命名空间提取出来 比如 base.MyData ，提取出 base, 要先判断一下有没有
					if strings.Contains(f.Type.Name, ".") {
						ns := strings.Split(f.Type.Name, ".")
						ns = ns[:len(ns)-1]
						//fieldNameSpace = strings.Join(ns, ".")
					}

					// Create expanded fields from the struct's fields
					for _, structField := range referencedStruct.Fields {
						// Create a new field with adjusted ID to avoid conflicts
						adjustedField := *structField
						// 使用字段原始ID * 1000 + 结构体字段ID作为偏移量，避免多个展开字段间的ID冲突
						adjustedField.ID = structField.ID + (f.ID * 1000) // 使用字段ID作为偏移量基础

						// 修复：处理嵌套引用类型，如 enums.ErrorCode
						// 对于展开字段，我们不需要调整类型名称，因为展开字段应该使用原始的类型名称
						// 类型解析将在 resolveTypesAndValues 阶段进行

						//log.Printf("展开： field %s with struct field %s", f.Name, structField.Name)
						// 为展开字段生成正确的方法名称，确保首字母大写
						expandedFieldName := st.scope.Add(common.UpperFirstRune(string(Name(structField.Name))), structField.Name)
						expandedField := &Field{
							Field:               &adjustedField,
							name:                Name(expandedFieldName),
							reader:              Name("ReadField" + id2str(adjustedField.ID)),
							writer:              Name("writeField" + id2str(adjustedField.ID)),
							getter:              Name("Get" + expandedFieldName),
							setter:              Name("Set" + expandedFieldName),
							isset:               Name("IsSet" + expandedFieldName),
							deepEqual:           Name("Field" + id2str(adjustedField.ID) + "DeepEqual"),
							isNested:            false,
							originalStructField: structField, // 设置原始结构体字段用于类型解析
							//namespace: fieldNameSpace, // 设置字段的命名空间
						}
						// 类型解析将在 resolveTypesAndValues 阶段进行

						expandedFields = append(expandedFields, expandedField)
					}

				}
			}
		}

		field := &Field{
			Field:          f,
			name:           Name(fn),
			reader:         Name(st.scope.Get(_p("read:" + id))),
			writer:         Name(st.scope.Get(_p("write:" + id))),
			getter:         Name(st.scope.Get(_p("get:" + f.Name))),
			setter:         Name(st.scope.Get(_p("set:" + f.Name))),
			isset:          Name(st.scope.Get(_p("isset:" + f.Name))),
			deepEqual:      Name(st.scope.Get(_p("deepequal:" + id))),
			isNested:       isNested,
			isExpandable:   isExpandable,
			expandedFields: expandedFields,
		}

		st.fields = append(st.fields, field)
	}

	if cu.Features().NoAliasTypeReflectionMethod && isAliasType(v) {
		st.isAlias = true
	}

	if len(usedName) > 0 {
		s.synthesized = append(s.synthesized, st)
	} else {
		ss := map[string]*[]*StructLike{
			"struct":    &s.structs,
			"union":     &s.unions,
			"exception": &s.exceptions,
		}[v.Category]
		if ss == nil {
			cu.Warn(fmt.Sprintf("struct[%s].category[%s]", st.Name, st.Category))
		} else {
			*ss = append(*ss, st)
		}
	}
	return st
}

func (s *Scope) resolveTypesAndValues(cu *CodeUtils) {
	resolver := NewResolver(s, cu)
	frugalResolver := NewFrugalResolver(s, cu)

	ff := make(chan *Field)

	go func() {
		ss := append(s.StructLikes(), s.synthesized...)
		for _, st := range ss {
			for _, f := range st.fields {
				ff <- f
			}
		}
		close(ff)
	}()

	ensureType := func(t TypeName, e error) TypeName {
		if e != nil {
			println(s.ast.Filename)
			println(e.Error())
			os.Exit(2)
		}
		return t
	}
	ensureCode := func(c Code, e error) Code {
		if e != nil {
			println(s.ast.Filename)
			println(e.Error())
			os.Exit(2)
		}
		return c
	}
	for f := range ff {
		v := f.Field
		f.typeName = ensureType(resolver.ResolveFieldTypeName(v))
		// This is used to set the real field name for nested struct, ex.
		// type T struct {
		// 	*Nested
		// }
		if cu.Features().EnableNestedStruct && isNestedField(f.Field) {
			name := f.typeName.Deref().String()
			if strings.Contains(name, ".") {
				names := strings.Split(name, ".")
				name = names[len(names)-1]
			}
			f.name = Name(name)
		}
		f.frugalTypeName = ensureType(frugalResolver.ResolveFrugalTypeName(v.Type))
		f.defaultTypeName = ensureType(resolver.GetDefaultValueTypeName(v))
		if f.IsSetDefault() {
			f.defaultValue = ensureCode(resolver.GetFieldInit(v))
		}

		// 处理展开字段的类型解析
		for _, expandedField := range f.expandedFields {
			// 使用调整后的字段进行类型解析，确保命名空间正确
			expandedField.typeName = ensureType(resolver.ResolveFieldTypeName(expandedField.Field))
			expandedField.frugalTypeName = ensureType(frugalResolver.ResolveFrugalTypeName(expandedField.Field.Type))
			expandedField.defaultTypeName = ensureType(resolver.GetDefaultValueTypeName(expandedField.Field))
			if expandedField.IsSetDefault() {
				expandedField.defaultValue = ensureCode(resolver.GetFieldInit(expandedField.Field))
			}
			// 如果类型解析失败，尝试手动构建类型名称
			if expandedField.typeName == "" && expandedField.Field.Type.Name != "" {
				// 如果调整后的类型名包含命名空间，直接使用它
				if strings.Contains(expandedField.Field.Type.Name, ".") {
					expandedField.typeName = TypeName(expandedField.Field.Type.Name)
				}
			}
		}
	}
	for _, t := range s.typedefs {
		t.typeName = ensureType(resolver.ResolveTypeName(t.Type)).Deref()
	}
	for _, v := range s.constants {
		v.typeName = ensureType(resolver.ResolveTypeName(v.Type))
		v.init = ensureCode(resolver.GetConstInit(v.Name, v.Type, v.Value))
	}

	for _, svc := range s.services {
		if svc.Extends == "" {
			continue
		}
		ref := svc.GetReference()
		if ref == nil {
			svc.base = s.Service(svc.Extends)
		} else {
			idx := ref.GetIndex()
			svc.base = s.includes[idx].Scope.Service(ref.GetName())
		}
	}

	for _, svc := range s.services {
		for _, fun := range svc.functions {
			// 处理函数参数类型解析
			for _, f := range fun.argType.fields {
				a := *f
				a.name = Name(fun.scope.Get(f.Name))
				// 确保参数类型正确解析，应用命名空间修复逻辑
				if f.Type != nil && strings.Contains(f.Type.Name, ".") {
					// 如果类型名包含命名空间，需要修复它
					parts := strings.Split(f.Type.Name, ".")
					typeName := parts[len(parts)-1]
					// 从当前作用域获取正确的命名空间
					correctNamespace := ""
					// 如果作用域中没有命名空间，尝试从 includes 中获取
					for _, inc := range s.includes {
						if inc != nil && inc.Scope != nil {
							//typeName 是不是在 inc 中
							if inc.Scope.globals.Get(typeName) != "" {
								// log.Printf("REQ调试: %s 在 %s 中", typeName, inc.PackageName)
								correctNamespace = inc.PackageName
								break
							}
						}
					}
					var fullTypeName string
					if correctNamespace == "" {
						fullTypeName = typeName
					} else {
						fullTypeName = correctNamespace + "." + typeName
					}
					// 直接使用完整的类型名
					a.typeName = ensureType(TypeName("*"+fullTypeName), nil)
				} else {
					// 普通类型解析
					a.typeName = ensureType(resolver.ResolveTypeName(f.Type))
				}
				fun.arguments = append(fun.arguments, &a)
			}
			if !fun.Oneway {
				fs := fun.resType.fields
				if !fun.Void {

					// 对于展开字段，需要找到原始的类型而不是展开后的字段类型
					if fs[0].isExpandable && fs[0].originalStructField != nil {
						// 使用原始字段的类型
						resolvedType, err := resolver.ResolveTypeName(fs[0].originalStructField.Type)
						fun.responseType = ensureType(resolvedType, err)
					} else {
						// 检查是否是 success 字段，如果是，使用原始的函数返回类型
						if fs[0].Name == "success" {
							// 使用原始的函数返回类型，而不是展开后的类型
							// 这里需要确保使用正确的类型

							// 如果 Type.Name 包含命名空间，需要修复它
							if strings.Contains(fs[0].Type.Name, ".") {
								parts := strings.Split(fs[0].Type.Name, ".")
								typeName := parts[len(parts)-1]
								// 从当前作用域获取正确的命名空间
								correctNamespace := ""
								// 如果作用域中没有命名空间，尝试从 includes 中获取
								for _, inc := range s.includes {
									if inc != nil && inc.Scope != nil {
										//typeName 是不是在 inc 中
										if inc.Scope.globals.Get(typeName) != "" {
											// log.Printf("调试: %s 在 %s 中", typeName, inc.PackageName)
											correctNamespace = inc.PackageName
											break
										}
									}
								}
								var fullTypeName string
								if correctNamespace == "" {
									fullTypeName = typeName // 构建完整的类型名，如 "test.User"
								} else {
									fullTypeName = correctNamespace + "." + typeName // 构建完整的类型名，如 "test.User"
								}

								// 直接使用完整的类型名
								fun.responseType = ensureType(TypeName("*"+fullTypeName), nil)
							} else {
								resolvedType, err := resolver.ResolveTypeName(fs[0].Type)
								fun.responseType = ensureType(resolvedType, err)
							}
						} else {
							// 如果不是 success 字段，可能是展开后的字段，需要找到原始类型
							resolvedType, err := resolver.ResolveTypeName(fs[0].Type)
							fun.responseType = ensureType(resolvedType, err)
						}
					}
					fs = fs[1:]
				}
				for _, f := range fs {
					t := *f
					t.name = Name(fun.scope.Get(f.Name))
					fun.throws = append(fun.throws, &t)
				}
			}
		}
	}
}

func isNestedField(f *parser.Field) bool {
	return annotationContainsTrue(f.Annotations, nestedAnnotation)
}

func isExpandField(f *parser.Field) bool {
	return annotationContainsTrue(f.Annotations, expandAnnotation)
}

func isAliasType(s *parser.StructLike) bool {
	return annotationContainsTrue(s.Annotations, aliasAnnotation)
}

func isRefInterfaceField(g *Scope, f *parser.Field) bool {
	return isRefInterfaceType(g, f.Type)
}

func annotationContainsTrue(annos parser.Annotations, anno string) bool {
	vals := annos.Get(anno)
	if len(vals) == 0 {
		return false
	}
	if len(vals) > 1 {
		log.Printf("[WARN] %s annotation has been set multiple values", anno)
		return false
	}
	if strings.EqualFold(vals[0], "true") {
		return true
	}

	return false
}

// isBaseType 添加辅助函数用于判断是否为基础类型
func isBaseType(typeName string) bool {
	baseTypes := map[string]bool{
		"bool":   true,
		"byte":   true,
		"i8":     true,
		"i16":    true,
		"i32":    true,
		"i64":    true,
		"double": true,
		"string": true,
		"binary": true,
	}

	// 移除可能的命名空间前缀后再判断
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		typeName = parts[len(parts)-1]
	}

	return baseTypes[typeName]
}

// 是否是引用我类型
func isHasNamespace(typeName string) bool {
	if strings.Contains(typeName, ".") {
		return true
	}
	return false
}
