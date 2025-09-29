namespace ts test.recursive

struct DepartmentInfo {
  1: i32 id
  2: string deptName
  3: list<DepartmentInfo> children
}

struct UserInfo {
  1: string id
  2: string name
  3: DepartmentInfo department
}
