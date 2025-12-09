namespace ts test.fields

// 测试生成 fields.ts 文件 - 使用 ts.gen_fields = "true"
struct MerchantSettingInfo {
  1: required string templateId
  2: optional string settingName
  3: optional double serviceFeeRate
  4: optional bool requireMandatoryDeduction
  5: optional bool collectOverdueFine
  6: optional string overdueFineCalcBase
  7: optional double overdueFineRate
  8: optional string overdueFineSettleTo
  9: optional bool collectWithdrawalPenalty
  10: optional string withdrawalPenaltyCalcBase
  11: optional double withdrawalPenaltyRate
  12: optional double minWithdrawalPenalty
  13: optional i32 withdrawalFeeFreeDays
  14: optional i32 autoCancelConfirmDays
  15: optional i32 contractChangeCloseDays
  16: optional i32 orderAgeMin
  17: optional i32 orderAgeMax
  18: optional bool merchantInitialReview
  19: optional bool platformReview
  20: optional bool isDefaultSetting
  21: optional string status
  22: optional string description
} (ts.gen_fields = "true")

// 测试生成 fields.ts 文件 - 使用自定义文件名
struct UserInfo {
  1: required string userId
  2: optional string userName
  3: optional string email
  4: optional i32 age
  5: optional string phone
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

