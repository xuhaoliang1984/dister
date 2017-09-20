package gluster

import (
    "g/net/ghttp"
    "g/encoding/gjson"
    "errors"
    "fmt"
    "reflect"
)

// Service 查询
func (this *NodeApiService) Get(r *ghttp.ClientRequest, w *ghttp.ServerResponse) {
    name := r.GetRequestString("name")
    if name == "" {
        if this.node.Service.Size() > 1000 {
            w.ResponseJson(0, "too large service size, need a service name to search", nil)
        } else {
            w.ResponseJson(1, "ok", this.node.getServiceMapForApi())
        }
    } else {
        sc := this.node.getServiceForApiByName(name)
        if sc != nil {
            w.ResponseJson(1, "ok", sc)
        } else {
            w.ResponseJson(0, "service not found", nil)
        }
    }
}

// Service 新增
func (this *NodeApiService) Put(r *ghttp.ClientRequest, w *ghttp.ServerResponse) {
    this.Post(r, w)
}

// Service 修改
func (this *NodeApiService) Post(r *ghttp.ClientRequest, w *ghttp.ServerResponse) {
    list := make([]ServiceConfig, 0)
    err  := gjson.DecodeTo(r.GetRaw(), &list)
    if err != nil {
        w.ResponseJson(0, "invalid data type: " + err.Error(), nil)
        return
    }
    // 数据验证
    for _, v := range list {
        err  = validateServiceConfig(&v)
        if err != nil {
            w.ResponseJson(0, err.Error(), nil)
            return
        }
    }
    // 提交数据到leader
    for _, v := range list {
        err  = this.node.SendToLeader(gMSG_API_SERVICE_SET, gPORT_REPL, gjson.Encode(v))
        if err != nil {
            w.ResponseJson(0, err.Error(), nil)
            return
        }
    }
    w.ResponseJson(1, "ok", nil)
}

// 验证Service提交参数
func validateServiceConfig(sc *ServiceConfig) error {
    for k, m := range sc.Node {
        commonError := errors.New(fmt.Sprintf("invalid config of service: %s, type: %s, node index: %d", sc.Name, sc.Type, k))
        switch sc.Type {
            case "pgsql": fallthrough
            case "mysql":
                host, _ := m["host"]
                port, _ := m["port"]
                user, _ := m["user"]
                pass, _ := m["pass"]
                name, _ := m["database"]
                if host == nil || port == nil || user == nil || pass == nil || name == nil ||
                    reflect.TypeOf(host).String() != "string" ||
                    reflect.TypeOf(port).String() != "string" ||
                    reflect.TypeOf(user).String() != "string" ||
                    reflect.TypeOf(pass).String() != "string" ||
                    reflect.TypeOf(name).String() != "string" {
                    return commonError
                }

            case "tcp":
                host, _ := m["host"]
                port, _ := m["port"]
                if host == nil || port == nil ||
                    reflect.TypeOf(host).String() != "string" ||
                    reflect.TypeOf(port).String() != "string" {
                    return commonError
                }

            case "web":
                url, _ := m["url"]
                if url == nil || reflect.TypeOf(url).String() != "string" {
                    return commonError
                }

            case "custom":
                script, _ := m["script"]
                if script == nil || reflect.TypeOf(script).String() != "string" {
                    return commonError
                }
        }
    }
    return nil
}

// Service 删除
func (this *NodeApiService) Delete(r *ghttp.ClientRequest, w *ghttp.ServerResponse) {
    list := make([]string, 0)
    err  := gjson.DecodeTo(r.GetRaw(), &list)
    if err != nil {
        w.ResponseJson(0, "invalid data type: " + err.Error(), nil)
        return
    }
    err  = this.node.SendToLeader(gMSG_API_SERVICE_REMOVE, gPORT_REPL, gjson.Encode(list))
    if err != nil {
        w.ResponseJson(0, err.Error(), nil)
    } else {
        w.ResponseJson(1, "ok", nil)
    }
}


