namespace go test
namespace ts api

include "base.thrift"
include "user.thrift"


// 服务测试
service UserService {
    user.UserResp getUser(1: user.UserReq req) (api.get = "/users/{id}")
    void createUser(1: user.User user) (api.post = "/users")
    list<user.User> getUsers() (api.get = "/users")
    user.User getUserById(1: string id) (api.get = "/users/{id}")
    void updateUser(1: user.User user) (api.put = "/users/{id}")
    void deleteUser(1: string id) (api.delete = "/users/{id}")
    list<user.User> searchUsers(1: string keyword, 2: i32 page, 3: i32 size) (api.get = "/users/search")
}

