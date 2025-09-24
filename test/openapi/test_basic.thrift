namespace go example

// 用户信息结构体
struct User {
    1: required i32 id,
    2: required string name,
    3: optional string email,
    4: optional i32 age,
}

// 用户状态枚举
enum UserStatus {
    ACTIVE = 1,
    INACTIVE = 2,
    SUSPENDED = 3,
}

// 用户查询请求
struct GetUserRequest {
    1: required i32 userId,
}

// 用户查询响应
struct GetUserResponse {
    1: required User user,
    2: required UserStatus status,
}

// 用户服务
service UserService {
    // 获取用户信息
    GetUserResponse getUser(1: GetUserRequest request),
    
    // 创建用户
    User createUser(1: User user),
    
    // 更新用户信息
    User updateUser(1: i32 userId, 2: User user),
    
    // 删除用户
    bool deleteUser(1: i32 userId),
}
