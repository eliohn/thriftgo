namespace ts test.expandable
include "base_ct.thrift"
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



// 手机号验证码登录请求结构体
struct PhoneLoginReq {
    1: required string phone // 手机号
    2: required string code // 每个 code 只能使用一次，重复使用会失效
    3: optional string merchant_id // 商户ID
    4: optional string inviter_id // 邀请人ID
    5: base_ct.CaptchaReq captcha_req
}
// 服务定义
service ExpandableService {
  void testBase(1: BaseStruct req)
  void testExpandable(1: ExpandableStruct req)
  void testExtended(1: ExtendedStruct req)
  void testPhoneLogin(1: PhoneLoginReq req ) (api.post="/api/login")
}
