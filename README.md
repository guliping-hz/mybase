# mybase

go基础库，把一些常用的方法或者是类，封装一下，方便其他项目调用。

1. [events](./events) 模拟node中的事件分发监听机制。
2. [net2](./net2) 封装了TCP socket和websocket。项目拿来直接写业务逻辑，不用再去写socket的逻辑了。
3. [report](./report) 封装了版署实名请求和游戏上报
4. [钉钉通知](./help-gin-dingd.go) 一键使用钉钉通知，监听服务器错误日志
5. [日志](./log.go) 由于日志中使用了chan。所以写完日志需要sleep等待一下（仅限测试）。