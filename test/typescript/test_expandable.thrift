namespace ts test.expandable

// 基础结构体
struct BaseStruct {
  1: required string name
  2: optional i32 age
}

// 可展开的结构体
struct ExpandableStruct {
  1: optional string description
  2: optional i32 priority
} (expandable = "true")

// 继承基础结构体的结构体
struct ExtendedStruct {
  1: required string id
  2: optional string email
} (expandable = "BaseStruct")

// 服务定义
service ExpandableService {
  void testBase(1: BaseStruct req)
  void testExpandable(1: ExpandableStruct req)
  void testExtended(1: ExtendedStruct req)
}
