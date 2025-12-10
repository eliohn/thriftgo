namespace ts test.fields

// 测试生成 fields.ts 文件 - 使用 ts.gen_fields = "true"
struct MerchantSettingInfo {
  1: required string templateId // 模板ID
  2: optional string settingName // 设置名称
  3: optional double serviceFeeRate // 服务费率 
  4: optional bool requireMandatoryDeduction // 是否需要强制扣除
  5: optional bool collectOverdueFine // 是否收集逾期罚款
  6: optional string overdueFineCalcBase // 逾期罚款计算基础
  7: optional double overdueFineRate // 逾期罚款率
  8: optional string overdueFineSettleTo // 逾期罚款结算到
  9: optional bool collectWithdrawalPenalty // 是否收集提现罚款
  10: optional string withdrawalPenaltyCalcBase // 提现罚款计算基础
  11: optional double withdrawalPenaltyRate // 提现罚款率
  12: optional double minWithdrawalPenalty
  13: optional i32 withdrawalFeeFreeDays // 提现免费天数    
  14: optional i32 autoCancelConfirmDays // 自动取消确认天数
  15: optional i32 contractChangeCloseDays // 合同变更关闭天数
  16: optional i32 orderAgeMin // 订单年龄最小值
  17: optional i32 orderAgeMax // 订单年龄最大值
  18: optional bool merchantInitialReview // 商户初始审核
  19: optional bool platformReview // 平台审核
  20: optional bool isDefaultSetting // 是否默认设置
  21: optional string status // 状态
  22: optional string description // 描述
} (ts.gen_fields = "true")

// 测试生成 fields.ts 文件 - 使用自定义文件名
struct UserInfo {
  1: required string userId // 用户ID
  2: optional string userName // 用户名
  3: optional string email // 邮箱
  4: optional i32 age // 年龄
  5: optional string phone // 手机号
} (ts.gen_fields = "generator/userinfo.fields.ts")

// 测试不生成 fields.ts 文件的结构体（没有注解）
struct ProductInfo {
  1: required string productId
  2: optional string productName
  3: optional double price
  4: optional i32 stock
}

// 测试生成 fields.ts 文件 - 使用自定义文件名（从路径中提取）
struct OrderInfo {
  1: required string orderId
  2: optional string orderStatus
  3: optional double totalAmount
  4: optional i64 createTime
} (ts.gen_fields = "orderinfo.fields.ts")

// 测试带展开字段的结构体
struct BaseInfo {
  1: optional string baseId
  2: optional string baseName
} (expandable = "true")

struct ExtendedInfo {
  1: required string id
  2: optional string name
  3: optional BaseInfo baseInfo (thrift.expand = "true")
} (ts.gen_fields = "true")

