
namespace ts domain.base

include "enums.thrift"
// 常量测试
const string DEFAULT_NAME = "Unknown"
const i32 MAX_AGE = 120


struct Base {
    1: enums.ErrorCode code // 错误码
    2: string msg   // 错误信息
}