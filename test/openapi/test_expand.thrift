namespace go example

// 基础用户信息
struct BaseUser {
    1: required i32 id,
    2: required string name,
    3: optional string email,
}

// 可展开的用户信息结构体
struct UserInfo {
    1: required i32 age,
    2: required string phone,
    3: optional string address,
} (expandable = "true")

// 用户结构体，展开 UserInfo
struct User {
    1: required BaseUser base,
    2: required UserInfo info (thrift.expand = "true"),
    3: optional string status,
}

// 用户查询请求
struct GetUserRequest {
    1: required i32 userId,
    2: optional bool includeDeleted,
}

// 用户查询响应，展开 User
struct GetUserResponse {
    1: required User user (thrift.expand = "true"),
    2: required string message,
}

// 用户服务
service UserService {
    // 获取用户信息
    GetUserResponse getUser(1: GetUserRequest request),
    
    // 创建用户
    User createUser(1: User user),
}
