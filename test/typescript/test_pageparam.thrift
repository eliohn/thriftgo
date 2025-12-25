include "base.thrift"
include "order.thrift"
namespace ts test.pageparam



service OrderService {
  // ================= 订单 =================
    // 获取普通订单列表
    order.ListNormalOrdersResp ListNormalOrders(1: order.ListNormalOrdersReq req) (api.get="/orders/normal")
}

