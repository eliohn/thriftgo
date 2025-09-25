struct CaptchaInfo {
    1: string data (go.tag = "form:\"data\" json:\"data\" query:\"data\"");
    2: i32 code_data;
    3: string msg_data;
}

struct GetCaptchaResp {
    2: CaptchaInfo data (thrift.expand = "true");
    3: i32 xx_data;
    4: string yy_data;
}