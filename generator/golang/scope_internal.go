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

	"github.com/cloudwego/thriftgo/extension/thrift_option"

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
				// Verify that the struct name matches the field type name
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
				if shouldExpand || structIsExpandable {
					isExpandable = true
					if strings.Contains(f.Type.Name, ".") {
						ns := strings.Split(f.Type.Name, ".")
						ns = ns[:len(ns)-1]
					}

					// Create expanded fields from the struct's fields
					for _, structField := range referencedStruct.Fields {
						// Create a new field with adjusted ID to avoid conflicts
						adjustedField := *structField
						adjustedField.ID = structField.ID + (f.ID * 1000)
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
							originalStructField: structField, // Set the original struct field for type resolution
							//namespace: fieldNameSpace, // Set the field's namespace
						}
						// Type resolution will be performed in the resolveTypesAndValues stage

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
	// Resolve the field types
	s.resolveFieldTypes(ff, resolver, frugalResolver, cu, ensureType, ensureCode)

	// Resolve the typedefs and constants
	s.resolveTypedefsAndConstants(resolver, ensureType, ensureCode)

	// After expanding the fields, check which packages are not used.
	s.checkUnusedPackagesAfterExpansion(cu)
	// The basic service of the parsing service
	s.resolveServiceBases()
	// Resolve the function types
	s.resolveFunctionTypes(resolver, ensureType)
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

// isBaseType adds a helper function to determine if it is a basic type
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
	if strings.Contains(typeName, ".") {
		parts := strings.Split(typeName, ".")
		typeName = parts[len(parts)-1]
	}

	return baseTypes[typeName]
}

// Whether it is a reference type
func isHasNamespace(typeName string) bool {
	if strings.Contains(typeName, ".") {
		return true
	}
	return false
}

// checkUnusedPackagesAfterExpansion checks which packages are not used after expanding fields
// This method checks if there are still references to the original packages after expanding fields, and marks them as unused if not
func (s *Scope) checkUnusedPackagesAfterExpansion(cu *CodeUtils) {
	// Collect all packages marked as used
	usedPackages := make(map[string]bool)
	for _, inc := range s.includes {
		if inc != nil && inc.Scope != nil {
			pkgName := s.includeIDL(cu, inc.Scope.ast)
			// If the package is not in libNotUsed, it means it's marked as used
			if _, exists := s.imports.libNotUsed[pkgName]; !exists {
				usedPackages[inc.PackageName] = true
			}
		}
	}

	// Check if these used packages still have actual references
	actualUsedPackages := s.collectActuallyUsedPackages()

	// Re-mark unused packages
	for _, inc := range s.includes {
		if inc != nil && inc.Scope != nil {
			pkgName := s.includeIDL(cu, inc.Scope.ast)
			// If the package is marked as used but has no actual references, re-mark it as unused
			if usedPackages[inc.PackageName] && !actualUsedPackages[inc.PackageName] {
				s.imports.libNotUsed[pkgName] = true
			}
		}
	}
}

// collectActuallyUsedPackages collects all actually used package names
// Determines package usage by checking the resolved type names of all fields, typedefs and constants
func (s *Scope) collectActuallyUsedPackages() map[string]bool {
	actualUsedPackages := make(map[string]bool)

	// Check all field resolved type names
	ss := append(s.StructLikes(), s.synthesized...)
	for _, st := range ss {
		for _, f := range st.fields {
			// Check normal field resolved type names
			if f.typeName != "" && strings.Contains(string(f.typeName), ".") {
				parts := strings.Split(string(f.typeName), ".")
				if len(parts) >= 2 {
					ns := strings.Join(parts[:len(parts)-1], ".")
					actualUsedPackages[ns] = true
				}
			}

			// Check expanded field resolved type names
			for _, expandedField := range f.expandedFields {
				if expandedField.typeName != "" && strings.Contains(string(expandedField.typeName), ".") {
					parts := strings.Split(string(expandedField.typeName), ".")
					if len(parts) >= 2 {
						ns := strings.Join(parts[:len(parts)-1], ".")
						actualUsedPackages[ns] = true
					}
				}
			}
		}
	}

	// Check all typedefs resolved type names
	for _, t := range s.typedefs {
		if t.typeName != "" && strings.Contains(string(t.typeName), ".") {
			parts := strings.Split(string(t.typeName), ".")
			if len(parts) >= 2 {
				ns := strings.Join(parts[:len(parts)-1], ".")
				actualUsedPackages[ns] = true
			}
		}
	}

	// Check all constants resolved type names
	for _, v := range s.constants {
		if v.typeName != "" && strings.Contains(string(v.typeName), ".") {
			parts := strings.Split(string(v.typeName), ".")
			if len(parts) >= 2 {
				ns := strings.Join(parts[:len(parts)-1], ".")
				actualUsedPackages[ns] = true
			}
		}
	}

	return actualUsedPackages
}

// resolveFieldTypes resolves all field types
func (s *Scope) resolveFieldTypes(ff chan *Field, resolver *Resolver, frugalResolver *FrugalResolver, cu *CodeUtils, ensureType func(TypeName, error) TypeName, ensureCode func(Code, error) Code) {
	for f := range ff {
		v := f.Field
		f.typeName = ensureType(resolver.ResolveFieldTypeName(v))
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

		s.resolveExpandedFields(f, resolver, frugalResolver, cu, ensureType, ensureCode)
	}
}

// resolveExpandedFields resolves expanded field types and package references
func (s *Scope) resolveExpandedFields(f *Field, resolver *Resolver, frugalResolver *FrugalResolver, cu *CodeUtils, ensureType func(TypeName, error) TypeName, ensureCode func(Code, error) Code) {
	for _, expandedField := range f.expandedFields {
		expandedField.typeName = ensureType(resolver.ResolveFieldTypeName(expandedField.Field))
		expandedField.frugalTypeName = ensureType(frugalResolver.ResolveFrugalTypeName(expandedField.Field.Type))
		expandedField.defaultTypeName = ensureType(resolver.GetDefaultValueTypeName(expandedField.Field))
		if expandedField.IsSetDefault() {
			expandedField.defaultValue = ensureCode(resolver.GetFieldInit(expandedField.Field))
		}
		if expandedField.typeName == "" && expandedField.Field.Type.Name != "" {
			if strings.Contains(expandedField.Field.Type.Name, ".") {
				expandedField.typeName = TypeName(expandedField.Field.Type.Name)
			}
		}
		if strings.Contains(expandedField.Field.Type.Name, ".") {
			parts := strings.Split(expandedField.Field.Type.Name, ".")
			if len(parts) >= 2 {
				ns := strings.Join(parts[:len(parts)-1], ".")
				for _, inc := range s.includes {
					for _, refInc := range inc.Scope.includes {
						if refInc != nil && refInc.Scope != nil && refInc.PackageName == ns {
							pkgName := s.includeIDL(cu, refInc.Scope.ast)
							s.imports.UseStdLibrary(pkgName)
						}
					}
				}
			}
		}
	}
}

// resolveTypedefsAndConstants resolves typedefs and constants types
func (s *Scope) resolveTypedefsAndConstants(resolver *Resolver, ensureType func(TypeName, error) TypeName, ensureCode func(Code, error) Code) {
	for _, t := range s.typedefs {
		t.typeName = ensureType(resolver.ResolveTypeName(t.Type)).Deref()
	}
	for _, v := range s.constants {
		v.typeName = ensureType(resolver.ResolveTypeName(v.Type))
		v.init = ensureCode(resolver.GetConstInit(v.Name, v.Type, v.Value))
	}
}

// resolveServiceBases resolves service base services
func (s *Scope) resolveServiceBases() {
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
}

// resolveFunctionTypes resolves function parameter and return value types
func (s *Scope) resolveFunctionTypes(resolver *Resolver, ensureType func(TypeName, error) TypeName) {
	for _, svc := range s.services {
		for _, fun := range svc.functions {
			s.resolveFunctionArguments(fun, resolver, ensureType)
			if !fun.Oneway {
				s.resolveFunctionResponse(fun, resolver, ensureType)
			}
		}
	}
}

// resolveFunctionArguments resolves function parameter types
func (s *Scope) resolveFunctionArguments(fun *Function, resolver *Resolver, ensureType func(TypeName, error) TypeName) {
	for _, f := range fun.argType.fields {
		a := *f
		a.name = Name(fun.scope.Get(f.Name))
		if f.Type != nil && strings.Contains(f.Type.Name, ".") {
			parts := strings.Split(f.Type.Name, ".")
			typeName := parts[len(parts)-1]
			correctNamespace := ""
			for _, inc := range s.includes {
				if inc != nil && inc.Scope != nil {
					if inc.Scope.globals.Get(typeName) != "" {
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
			a.typeName = ensureType(TypeName("*"+fullTypeName), nil)
		} else {
			a.typeName = ensureType(resolver.ResolveTypeName(f.Type))
		}
		fun.arguments = append(fun.arguments, &a)
	}
}

// resolveFunctionResponse resolves function return value types
func (s *Scope) resolveFunctionResponse(fun *Function, resolver *Resolver, ensureType func(TypeName, error) TypeName) {
	fs := fun.resType.fields
	if !fun.Void {
		if fs[0].isExpandable && fs[0].originalStructField != nil {
			resolvedType, err := resolver.ResolveTypeName(fs[0].originalStructField.Type)
			fun.responseType = ensureType(resolvedType, err)
		} else {
			if fs[0].Name == "success" {
				if strings.Contains(fs[0].Type.Name, ".") {
					parts := strings.Split(fs[0].Type.Name, ".")
					typeName := parts[len(parts)-1]
					correctNamespace := ""
					for _, inc := range s.includes {
						if inc != nil && inc.Scope != nil {
							if inc.Scope.globals.Get(typeName) != "" {
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
					fun.responseType = ensureType(TypeName("*"+fullTypeName), nil)
				} else {
					resolvedType, err := resolver.ResolveTypeName(fs[0].Type)
					fun.responseType = ensureType(resolvedType, err)
				}
			} else {
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
