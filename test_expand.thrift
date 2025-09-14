namespace go test

const string VERSION = "1.0.0"


struct Base {
  1: i32 code
  2: string msg
} (expandable = "true", version = "1.0.0")

struct UserDTO {
  1: i64 id
  2: string name
  3: string email
}

struct UserResp {
  1: Base base
  3: UserDTO data
}
