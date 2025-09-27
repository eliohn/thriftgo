namespace ts expreq

struct PageReq{
  1:i32 pageNum (api.query = "page")
  2:i32 pageSize
} (expandable = "true")

struct TaskReq{
  1:string taskId
  2:PageReq pageReq
}

service Expreq {
  i32 add(1:TaskReq a) (api.get = "/add")
}
