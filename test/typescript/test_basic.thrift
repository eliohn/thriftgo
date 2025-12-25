namespace ts common.base


struct TestStruct {
  1: i32 a;
  2: string b;
}

struct TestStruct2 {
  1: i32 a;
  2: string b;
} (expandable = "true")

struct TestResp{
  1: TestStruct2 a;
  2: TestStruct b;
}

service TestService {
  TestResp test(1: i32 a, 2: string b);
}

// 分页请求结构体
struct PageReq {
  1: i32 page_num (api.query="page")  // 页码
  2: i32 page_size (api.query="pageSize")  // 每页大小
} (expandable = "true")