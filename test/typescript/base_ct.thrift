namespace go base.com
namespace ts base.ct

enum StatusCode {
    SUCCESS = 0,
    ERROR = 1,
}

struct BaseResp {
    1: required string StatusMessage = "",
    2: required i32 StatusCode = 0,
    3: optional map<string, string> Extra,
}

struct PingReq {
    1: required string message,
    2: required i32 code(api.path = "code"),
    3: required bool success,
}

service BaseService {
    BaseResp Ping(1: PingReq req),
}