struct CaptchaInfo {
    1: string data (go.tag = "form:\"data\" json:\"data\" query:\"data\"");
}

struct GetCaptchaResp {
    1001: i32 code;
    1002: string msg;
    2: CaptchaInfo data (thrift.expand = "true");
}