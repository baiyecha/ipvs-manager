# 说明
## 启动方式
### 集群启动
server端
```
./ipvs-manager  --raft_node_id node0 --adverties 192.168.143.77:8110 --cluster 192.168.143.77:8010,192.168.143.78:8010,192.168.143.80:8010
./ipvs-manager  --raft_node_id node1 --adverties 192.168.143.78:8110 --cluster 192.168.143.77:8010,192.168.143.78:8010,192.168.143.80:8010
./ipvs-manager  --raft_node_id node2 --adverties 192.168.143.80:8110 --cluster 192.168.143.77:8010,192.168.143.78:8010,192.168.143.80:8010
```
agent 端
```
ipvs-manage --server_type agent --grpc_address 192.168.143.77:8210,192.168.143.78:8210,192.168.143.80:8210
```
参数说明
```
[root@raft0 tuzhigen]# ./ipvs-manager -h
Usage of ./ipvs-manager:
      --adverties string      集群raft广播出来的地址，集群之间用这个地址通信 (default "127.0.0.1:8110")
      --cluster string        集群所有节点的http地址，用来对接raft (default "127.0.0.1:8110")
      --dummy_name string     ipvs dummy网卡的名字 (default "ipvs-manager")
      --grpc_address string   agent对接的grpc地址列表 (default "127.0.0.1:8210")
      --grpc_port int         grpc的监听地址 (default 8210)
      --raft_node_id string   raft 的节点id,每个节点需要保持唯一 (default "raft")
      --raft_vol_dir string   raft⋅信息和kv数据库的文件目录 (default "node_1_data")
      --server_port int       http的端口服务 (default 8010)
      --server_type string    启动方式，默认是只启动server服务编辑ipvs策略, singleon 为all-in-one模式，agent为部署agent控制ipvs
```
使用-h 查看各个参数说明

### 单点启动
```
ipvs-manage --server_type singleon
```
使用singleon启动后，节点上会同时运行一个server和agent，参数都为默认配置

## raft 模块说明
1. server
server模块提供一个http服务，提供两种接口
a. 数据库的增删查改和web界面
b. raft集群的管理接口

2. fsm
fsm 被称作有限状态机，是写入数据的一个具体的实现，在golang里面，他是一个interface，需要自己实现一套
```
/*FSM provides an interface that can be implemented by
clients to make use of the replicated log.*/
type FSM interface {
    /* Apply log is invoked once a log entry is committed.
    It returns a value which will be made available in the
    ApplyFuture returned by Raft.Apply method if that
    method was called on the same Raft node as the FSM.*/
    Apply(*Log) interface{}
    // Snapshot is used to support log compaction. This call should
    // return an FSMSnapshot which can be used to save a point-in-time
    // snapshot of the FSM. Apply and Snapshot are not called in multiple
    // threads, but Apply will be called concurrently with Persist. This means
    // the FSM should be implemented in a fashion that allows for concurrent
    // updates while a snapshot is happening.
    Snapshot() (FSMSnapshot, error)
    // Restore is used to restore an FSM from a snapshot. It is not called
    // concurrently with any other command. The FSM must discard all previous
    // state.
    Restore(io.ReadCloser) error
}
```
1. Apply 方法我们采用使用内嵌的kv数据库badger实现，持久化到磁盘，这里也不是直接写入，而是使用raft的Apply方式，为这次set操作生成一个log entry，这里面会根据raft的内部协议，在各个节点之间进行通信协作，确保最后这条log 会在整个集群的节点里面提交或者失败。
2. Snapshot 快照，因为我们直接使用的持久化kv数据库，已经存盘保存了，所以不需要使用到快照功能
3. Restore 服务重启的时候，会先读取本地的快照来恢复数据，在FSM里面定义的Restore函数会被调用，我们将数据写入到kv数据库里面即可。

### 流程说明
1. 集群启动的时候，会根据cluster参数，遍历接口，找到leader，如果没有找到，则自身成为leader运行，如果找到了leader，则自身发送http请求，申请自己加入到集群中
2. 先启动的节点收到请求后，获取对方的地址（指raft集群内部通信的tcp地址），然后调用AddVoter把这个节点加入到集群即可。申请加入的节点会进入follower状态，这以后集群节点之间就可以正常通信，leader也会把数据同步给follower。

## healthCheck 逻辑说明
1. server端启动之后，会开启一个forever的协程，用来检查ipvs的后端服务是否正常，如果不正常，将状态改为不健康
2. 协程启动后，会先检查自身是否是leader节点，不是leader节点不进行检查，健康检查由leader负责
3. 健康检查提供tcp检查和http检查方式，http必须指定检查的url，tcp默认使用ipvs后端的ip + port

## agent 端逻辑
1. 类似kube-proxy，通过grpc定时获取ipvs的配置信息，来对比本机的ipvs配置(只对比ipvs网卡上的),将本机的ipvs修改成为用户期望的配置。
2. 在创建ipvs的过程中，会增加一个dummy类型的网卡，并在该网卡上绑定ipvs的vip，增加snat规则
3. 对比发现ipvs的后端或者配置发生变化，就更新整个ipvs规则。如果发现ipvs需要删除，那就只删除ipvs的规则，保留ipvs的网卡和vip
### ipvs 创建逻辑
1. 创建dummy类型的网卡，dummy是一个虚拟的网卡，类似回环网络，因为ipvs工作在input链上，所以需要将请求导入到本机的网络栈。
2. 网卡配置完成之后，将ipvs的vip绑定到该网卡上， 类似 ip addr add 1.1.1.2 dev dummy
3. ipvs会根据规则，修改请求包的源地址和目标地址，目标地址修改ipvs后端的ip，源地址为1.1.1.2，回包的时候会有问题，所以需要进行snat，将1.1.1.2转换为物理网卡的ip。
4. 根据用户配置的ipvs规则，调用代码创建ipvs。

## 参考资料
* https://zhuanlan.zhihu.com/p/58048906
* https://yusufs.medium.com/creating-distributed-kv-database-by-implementing-raft-consensus-using-golang-d0884eef2e28
