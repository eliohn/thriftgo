namespace ts test.defaults

// 空结构体
struct EmptyStruct {
}

struct InnerStruct {
  1: optional string name22
  2: optional i32 age22
} (expandable = "true")

// 所有字段都是可选的结构体
struct OptionalStruct {
  1: optional string name
  2: optional i32 age
  3: optional InnerStruct innerStruct
}

// 有必需字段的结构体
struct RequiredStruct {
  1: required string name
  2: optional i32 age
}

service TestService {
  // 测试空结构体参数
  void testEmpty(1: EmptyStruct empty)
  
  // 测试所有字段都是可选的结构体参数
  void testOptional(1: OptionalStruct optional)
  
  // 测试有必需字段的结构体参数
  void testRequired(1: RequiredStruct required)
  
  // 测试混合参数
  void testMixed(1: EmptyStruct empty, 2: OptionalStruct optional, 3: RequiredStruct required)
}
