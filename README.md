# TOWGO 轻量级开源分布式应用开发框架


towgo 是由go语言开发实现的轻量级应用服务框架。框架语法简洁易懂，编写一套控制器逻辑即可以运行在不同的应用层服务（HTTP、WEBSOCKET、TCP）
towgo 不再是传统的web服务开发框架,更是一套应用服务的开发框架。
towgo 天生支持微服务架构，包含了微服务常用的基本的组件（网关路由、边缘节点、边车路由等）
towgo 摒弃了传统的http应用层协议作为接口交互通讯，换成了jsonrpc协议。但这并不意味着框架不能用于web开发，反而可以提高开发效率，这是因为通过rpc控制器开发的接口不仅可以暴露在http接口上，也可以暴露在websocket、tcp等接口上，从而实现一次开发即可支持多个底层通讯协议。这样的模式不仅可以用于对外的传统的http服务接口,内部服务的相互调用，可以使用性能更为高效的websocket、tcp作为传输协议。应用范围相当广泛（互联网、物联网、智能设备、游戏服务端等）

- 简洁易用
- 分布式系统支持
- 同时支持多种ORM引擎 （GORM,XORM）
- 支持读写分离
- 一行代码实现增删改查接口
- 丰富的模块
- 丰富的DEMO
- 信创国产化可用框架
```
jsonrpc2/                      // 框架根包名（对外暴露的核心包）
├── client/                    // 客户端模块（发起RPC调用）
│   ├── client.go              // 客户端核心逻辑（Call/Notify/BatchCall）
│   ├── options.go             // 客户端配置选项（超时、传输协议等）
│   └── internal/              // 客户端内部实现（对外不暴露）
│       └── caller.go          // 调用逻辑封装
├── server/                    // 服务端模块（处理RPC请求）
│   ├── server.go              // 服务端核心逻辑（启动/停止/注册方法/注册拦截器）
│   ├── options.go             // 服务端配置选项（监听地址、传输协议等）
│   └── internal/              // 服务端内部实现（对外不暴露）
│       └── dispatcher.go      // 请求分发逻辑（匹配方法+执行拦截器链）
├── protocol/                  // 核心协议层（纯JSON-RPC 2.0规范实现，与传输无关）
│   ├── consts.go              // 协议常量（标准错误码、jsonrpc版本等）
│   ├── request.go             // 请求对象定义+校验逻辑
│   ├── response.go            // 响应对象定义+封装逻辑
│   ├── batch.go               // 批量请求/响应处理逻辑
│   └── validator.go           // 协议格式校验工具
├── interceptor/               // 拦截器层（GoFrame风格）
│   ├── chain.go               // 拦截器链（责任链模式核心）
│   ├── interceptor.go         // 拦截器接口定义+基础类型
│   └── builtin/               // 内置拦截器（日志/鉴权/超时等，开箱即用）
│       ├── logger.go          // 日志拦截器
│       ├── auth.go            // 鉴权拦截器
│       └── timeout.go         // 超时控制拦截器
├── context/                   // 上下文层（全流程数据载体）
│   ├── context.go             // Context接口定义
│   ├── basic_context.go       // 基础上下文实现（通用场景）
│   └── value.go               // 上下文键值对操作工具
├── transport/                 // 传输适配层（多协议实现）
│   ├── transport.go           // Transport接口定义（统一所有传输协议）
│   ├── http/                  // HTTP传输实现
│   │   ├── server.go          // HTTP服务端（实现Transport接口）
│   │   └── client.go          // HTTP客户端（实现Transport接口）
│   ├── websocket/             // WebSocket传输实现
│   │   ├── server.go
│   │   ├── client.go
│   │   └── conn.go            // WebSocket连接管理
│   ├── tcp/                   // TCP传输实现
│   │   ├── server.go
│   │   ├── client.go
│   │   └── codec.go           // TCP粘包/拆包编解码器
│   └── udp/                   // UDP传输实现
│       ├── server.go
│       └── client.go
├── util/                      // 基础工具层（通用能力）
│   ├── reflect.go             // 反射工具（解析params参数、方法调用）
│   ├── codec.go               // JSON编解码工具
│   ├── logger.go              // 日志工具（适配不同日志库）
│   └── sync.go                // 并发安全工具（锁、连接池等）
├── errors/                    // 统一错误处理（封装标准错误+自定义错误）
│   ├── errors.go              // 错误类型定义+构造函数
│   └── code.go                // 错误码常量（协议码+业务码）
├── example/                   // 示例代码（用户参考）
│   ├── http_server.go         // HTTP服务端示例
│   ├── websocket_client.go    // WebSocket客户端示例
│   └── tcp_batch_call.go      // TCP批量请求示例
├── test/                      // 测试目录（与源码结构对应）
│   ├── protocol/
│   ├── interceptor/
│   └── transport/
├── go.mod                     // Go模块配置
├── go.sum
└── README.md                  // 框架文档（使用说明、架构说明）
```
