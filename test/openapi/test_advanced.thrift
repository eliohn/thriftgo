namespace go example

// 用户信息结构体
struct User {
    1: required i32 id,
    2: required string name,
    3: optional string email,
    4: optional i32 age,
    5: optional UserStatus status,
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
    2: optional bool includeDeleted,
}

// 用户查询响应
struct GetUserResponse {
    1: required User user,
    2: required UserStatus status,
    3: optional string message,
}

// 用户列表请求
struct ListUsersRequest {
    1: optional i32 page,
    2: optional i32 pageSize,
    3: optional string keyword,
}

// 用户列表响应
struct ListUsersResponse {
    1: required list<User> users,
    2: required i32 total,
    3: required i32 page,
    4: required i32 pageSize,
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
    
    // 获取用户列表
    ListUsersResponse listUsers(1: ListUsersRequest request),
}

// 订单服务
service OrderService {
    // 创建订单
    Order createOrder(1: CreateOrderRequest request),
    
    // 获取订单
    Order getOrder(1: i64 orderId),
    
    // 更新订单状态
    bool updateOrderStatus(1: i64 orderId, 2: OrderStatus status),
}

// 订单结构体
struct Order {
    1: required i64 id,
    2: required i32 userId,
    3: required list<OrderItem> items,
    4: required double totalAmount,
    5: required OrderStatus status,
    6: required i64 createdAt,
}

// 订单项
struct OrderItem {
    1: required i32 productId,
    2: required string productName,
    3: required i32 quantity,
    4: required double price,
}

// 订单状态
enum OrderStatus {
    PENDING = 1,
    CONFIRMED = 2,
    SHIPPED = 3,
    DELIVERED = 4,
    CANCELLED = 5,
}

// 创建订单请求
struct CreateOrderRequest {
    1: required i32 userId,
    2: required list<OrderItem> items,
}
