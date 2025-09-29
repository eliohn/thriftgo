namespace ts common.base

struct Empty {
}

struct NilResp {
  1: string message
}

namespace ts admin.structs

struct GetSystemConfigResp {
  1: string config
}

struct UpdateSystemConfigReq {
  1: string config
}

struct ClearCacheReq {
  1: string key
}

struct HealthCheckResp {
  1: string status
}

namespace ts admin.service

service AdminService {
  GetSystemConfigResp getSystemConfig()
  void updateSystemConfig(1: UpdateSystemConfigReq req)
  void clearCache(1: ClearCacheReq req)
  HealthCheckResp healthCheck()
}