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