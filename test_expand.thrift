namespace go test

include "base.thrift"


struct UserDTO {
  1: i64 id
  2: string name
  3: string email
}

struct UserResp {
  1: base.Base base
  3: UserDTO data
}
