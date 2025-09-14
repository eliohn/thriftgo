namespace go base

enum ErrorCode {
    Success = 0,
    Unknown = 1,
    InvalidParam = 2,
}

struct Base {
    1: ErrorCode code
    2: string msg
} (expandable = "true", version = "1.0.0")


struct MyData {
    1: i64 id
    2: string name
    3: string email

}