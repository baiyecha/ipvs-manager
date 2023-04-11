# 说明
## 启动方式
```
SERVER_PORT=2221 RAFT_NODE_ID=node1 RAFT_PORT=1111 RAFT_VOL_DIR=node_1_data ./ipvs-manager
SERVER_PORT=2222 RAFT_NODE_ID=node2 RAFT_PORT=1112 RAFT_VOL_DIR=node_2_data RAFT_LEADER=127.0.0.1:2221 ./ipvs-manager
SERVER_PORT=2223 RAFT_NODE_ID=node3 RAFT_PORT=1113 RAFT_VOL_DIR=node_3_data RAFT_LEADER=127.0.0.1:2221 ./ipvs-manager
```
参数说明
* SERVER_POR http的端口
* RAFT_NODE_ID 集群中的id,保持唯一性就行
* RAFT_VOL_DIR raft文件和数据库文件持久化的目录
* RAFT_LEADER 当节点作为个新的成员加入到集群中时，需要指定leader的http服务地址，启动后会自动join到集群中，记住是新的节点，重启的节点由于已经保留了集群的信息，不需要该参数

## raft 模块说明
1. server
server模块提供一个http服务，提供两种接口
a. 数据库的增删查改和web界面
b. raft集群的管理接口

2. fsm
fsm 被称作有限状态机，是写入数据的一个具体的实现，在golang里面，他说第一个interface，需要自己实现一套
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

## 流程说明
1. 集群最开始的时候只有一个节点，我们让第一个节点通过bootstrap的方式启动，它启动后成为leader。
2. 后续的节点启动的时候需要加入集群，启动的时候指定第一个节点的地址，并发送请求加入集群，这里我们定义成直接通过http请求。
3. 先启动的节点收到请求后，获取对方的地址（指raft集群内部通信的tcp地址），然后调用AddVoter把这个节点加入到集群即可。
申请加入的节点会进入follower状态，这以后集群节点之间就可以正常通信，leader也会把数据同步给follower。

## 参考资料
* https://zhuanlan.zhihu.com/p/58048906
* https://yusufs.medium.com/creating-distributed-kv-database-by-implementing-raft-consensus-using-golang-d0884eef2e28
