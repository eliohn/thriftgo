namespace go test.enum.color
namespace ts test.enum.color

// 测试枚举 color 功能
enum Status {
  ACTIVE = 1 (ts.tag = "激活", ts.color = "#52c41a"),
  INACTIVE = 2 (ts.tag = "未激活", ts.color = "#ff4d4f"),
  PENDING = 3 (ts.tag = "待处理", ts.color = "#faad14")
}

// 测试只有 tag 没有 color 的枚举
enum Priority {
  LOW = 1 (ts.tag = "低"),
  MEDIUM = 2 (ts.tag = "中"),
  HIGH = 3 (ts.tag = "高")
}

// 测试混合情况
enum TaskType {
  BUG = 1 (ts.tag = "Bug", ts.color = "#ff4d4f"),
  FEATURE = 2 (ts.tag = "功能"),
  IMPROVEMENT = 3 (ts.tag = "改进", ts.color = "#1890ff")
}
