
namespace ts domain.uservo

// 基本类型测试
struct User {
    1: required string name
    2: optional i32 age
    3: required bool active
    4: optional double score
    5: optional string id (api.path = "id")
    6: optional string page (api.query = "page")
    7: optional string size (api.query = "pageSize")
}

struct UserReq {
    1: required string name (api.path = "name")
    2: optional i32 age
}

struct UserResp {
    1: required User user
    2: optional string message
}