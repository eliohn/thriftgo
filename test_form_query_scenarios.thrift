namespace go test

struct TestStruct {
    // 场景1: 没有任何 go.tag 注解 - 应该自动添加 form 和 query 标签
    1: string page_resp;
    
    // 场景2: 只有 form 标签 - 应该自动添加 query 标签
    2: optional i32 page_size (go.tag = 'form:"page_size"');
    
    // 场景3: 只有 query 标签 - 应该自动添加 form 标签  
    3: required string user_name (go.tag = 'query:"user_name"');
    
    // 场景4: 同时有 form 和 query 标签 - 不应该添加任何标签
    4: optional bool is_active (go.tag = 'form:"is_active" query:"is_active"');
    
    // 场景5: 有其他标签但没有 form 和 query - 应该添加 form 和 query 标签
    5: optional string description (go.tag = 'json:"description"');
}
