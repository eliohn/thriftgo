namespace go test.enum.tag
namespace ts test.enum.tag

// 测试枚举 tag 功能
enum MenuType {
  DIRECTORY = 1 (ts.tag = "目录"),
  MENU = 2 (ts.tag = "菜单"),
  BUTTON = 3 (ts.tag = "按钮")
}

// 测试枚举 color 功能
enum Status {
  ACTIVE = 1 (ts.tag = "激活", ts.color = "#52c41a"),
  INACTIVE = 2 (ts.tag = "未激活", ts.color = "#ff4d4f"),
  PENDING = 3 (ts.tag = "待处理", ts.color = "#faad14")
}

// 测试部分有 tag 的枚举
enum Priority {
  LOW = 1 (ts.tag = "低"),
  MEDIUM = 2,
  HIGH = 3 (ts.tag = "高")
}