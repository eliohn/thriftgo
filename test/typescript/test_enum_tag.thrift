namespace go test.enum.tag
namespace ts test.enum.tag

// 测试枚举 tag 功能
enum MenuType {
  DIRECTORY = 1 (ts.tag = "目录"),
  MENU = 2 (ts.tag = "菜单"),
  BUTTON = 3 (ts.tag = "按钮")
}

// 测试没有 tag 的枚举
enum Status {
  ACTIVE = 1,
  INACTIVE = 2,
  PENDING = 3
}

// 测试部分有 tag 的枚举
enum Priority {
  LOW = 1 (ts.tag = "低"),
  MEDIUM = 2,
  HIGH = 3 (ts.tag = "高")
}