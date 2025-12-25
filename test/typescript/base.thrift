namespace ts common.base

// 分页请求结构体
struct PageReq {
  1: i32 page_num (api.query="page")  // 页码
  2: i32 page_size (api.query="pageSize")  // 每页大小
} (expandable = "true")