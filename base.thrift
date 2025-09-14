namespace go base

include "enums.thrift"

struct Base {
    1: enums.ErrorCode code
    2: string msg
} (expandable = "true")


struct MyData {
    1: i64 id
    2: string name
    3: string email

}