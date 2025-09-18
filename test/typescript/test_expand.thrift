namespace ts test.expand

struct BaseInfo {
  1: string name
  2: i32 age
}

struct UserInfo {
  1: BaseInfo base (thrift.expand="true")
  2: string email
}
