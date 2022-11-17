# node-collector

k8s集群node节点检测collector

## 环境配置

go1.18+版本
* prometheus需要go1.17+
* cilium/ebpf需要go1.18+


## 运行

运行`run.sh`脚本即可

注意2点：

*   tip1: 请先关闭可能存在的旧的collector进程，否则相关端口被占用，http服务运行不起来会直接退出
    可行做法：ps aux | grep ./collector  + sudo kill -9 (因为旧的collector是用root运行的)
*   tip2: 建议在编译阶段不要用 sudo，比如 sudo go run *.go 相当于编译阶段也是在sudo中进行
    ridx/k8s 和 root 用户是两套go环境，root构建需要重新拉外部module，因此还需要先换goproxy。建议先用普通用户go build，最后再用sudo执行可执行文件，也即`run.sh`中的做法

## 升级

在本地仓库 pull 代码（上游已经设置为 `https://github.com/yuqi-lee/node-collector.git`）

根据需要修改本地的 `config.json` 中的内容



