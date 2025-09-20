
namespace go thrift.test

enum TestEnum {
  TEST_VALUE_1 = 1,
  TEST_VALUE_2 = 2
}
struct PageResp {
  1: required i32 page_num
  2: required i32 page_size
}


struct TestStruct {
  1: required string Name
  2: PageResp Page (thrift.expand="true")
}
