include "base_ct.thrift"

namespace ts patch_test

struct UpdateUserReq {
  1: string userId (api.path = "userId")
  2: string name (api.body = "name")
  3: string email (api.body = "email")
  4: i32 age (api.body = "age")
}

struct UpdateUserResp {
  1: base_ct.BaseResp baseResp
  2: string userId
  3: string name
  4: string email
  5: i32 age
}

struct PartialUpdateReq {
  1: string id (api.path = "id")
  2: optional string name (api.body = "name")
  3: optional string email (api.body = "email")
  4: optional i32 age (api.body = "age")
}

struct PartialUpdateResp {
  1: base_ct.BaseResp baseResp
  2: string id
  3: optional string name
  4: optional string email
  5: optional i32 age
}

service UserService {
  // 使用 PATCH 方法进行部分更新
  PartialUpdateResp updateUser(1: PartialUpdateReq req) (api.patch = "/users/:id")
  
  // 使用 PUT 方法进行完整更新
  UpdateUserResp replaceUser(1: UpdateUserReq req) (api.put = "/users/:userId")
  
  // 使用 GET 方法获取用户信息
  UpdateUserResp getUser(1: string userId) (api.get = "/users/:userId")
  
  // 使用 DELETE 方法删除用户
  base_ct.BaseResp deleteUser(1: string userId) (api.delete = "/users/:userId")
}
