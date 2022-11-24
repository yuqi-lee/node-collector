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

> git pull --rebase

各个host上本地仓库路径：

* skv-node1: /home/k8s/exper/lyq/node-collector
* skv-node3: /home/k8s/exper/lyq/node-collector
* skv-node4: /home/ridx/lyq/k8s/node-collector

**整个进程的配置信息依赖于node-collector目录下的`config.json`文件** （这个文件没有上传到github仓库，各个host根据需要自己进行维护）
根据需要修改本地的 `config.json` 中的内容


## 源代码文件

* `exporter.go`: 定义`prometheus`数据类型、`monitior`类函数（用于持续调用`record`类函数）、整个进程的初始化（读配置文件、初始化podname到ip的映射关系等），以及`main`函数
* `metrics.go`: 定义`record`类函数:一个`record`函数只一次性记录一个数据或一组相似数据。**`record`函数的调用时间间隔即为这类指标打点的时间间隔**
* `bpf.go`: 定义读取`bpf map`的接口值类型
* `config.go`： 定义`collector`的配置信息及初始化方法
* `podinfo.go`: 定义获取pod信息的方法，以及在内存中更新pod信息的方法
* `utils.go`: 一些工具类函数，int转ip、截取字符串等
* `*_test.go`: 单元测试