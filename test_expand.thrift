namespace go test

include "base.thrift"

enum BtextCode {
  Success = 0,
  Unknown = 1,
  InvalidParam = 2,
}

struct User {
  1: i64 id
  2: string name
  3: string email
  4: BtextCode phone
}
struct UserDTO {
  1: i64 id
  2: string name
  3: string email
}

struct UserResp {
  1: base.Base base
  3: UserDTO data
  4: User user
  5: base.MyData myData
}
