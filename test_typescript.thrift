namespace go test
namespace ts test_temp

include "base.thrift"

// 基本类型测试
struct User {
    1: required string name
    2: optional i32 age
    3: required bool active
    4: optional double score

}
struct UserReq {
    1: required string name
    2: optional i32 age
}

struct UserResp {
    1: required base.Base base
    2: User  user
}
// 枚举测试
enum Status {
    PENDING = 0
    ACTIVE = 1
    INACTIVE = 2
}

// 服务测试
service UserService {
    UserResp getUser(1: UserReq id) (api.get = "/users/{id.name}")
    void createUser(1: User user) (api.post = "/users")
    list<User> getUsers() (api.get = "/users")
}

// 常量测试
const string DEFAULT_NAME = "Unknown"
const i32 MAX_AGE = 120

// 类型定义测试
typedef string UserID
typedef map<string, User> UserMap
