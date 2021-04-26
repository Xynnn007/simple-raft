# simple-raft

基于Raft的分布式框架

## 节点配置

位于`test/peers.toml`下的是节点配置文件格式

## 启动参数

`-v` 获取版本信息
`-p <toml路径>` 指定toml路径，默认会在`./test/peers.toml`下寻找

## Make脚本

`make`执行可以在`bin`目录下编译出main文件

## 测试方法

`build`目录下包括有三个节点的启动脚本，即`run1.sh`，`run2.sh`和`run3.sh`，分别在三个终端下启动，
则会分别在8081、8082、8083三个端口启动服务，网络层采用http协议。