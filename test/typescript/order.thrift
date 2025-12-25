include "base.thrift"

namespace ts domain.orderVO

// 获取普通订单列表请求
struct ListNormalOrdersReq {
    1: optional string merchant_id                  // 商户ID
    2: optional string order_no                      // 订单编号
    3: optional string customer_name                 // 客户名称
    4: optional string salesperson                  // 业务员
    5: optional string start_time                   // 开始时间
    6: optional string end_time                     // 结束时间
    7: optional i32 status  // 订单状态
    8: optional base.PageReq page_param             // 分页参数
}
// 获取普通订单列表响应
struct ListNormalOrdersResp {
    1: required string message
}