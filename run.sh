# tip1: 请先关闭可能存在的旧的collector进程，否则相关端口被占用，http服务运行不起来会直接退出
# 可行做法：ps aux | grep collector  + sudo kill -9 (因为旧的collector是用root运行的)

# tip2：建议在编译阶段不要用 sudo，比如 sudo go run *.go 相当于 编译阶段也在sudo中做了
# ridx/k8s 和 root 用户是两套go环境，root构建需要重新拉外部module，因此还需要先换goproxy
# 因此建议先用普通用户go build，最后再用sudo执行可执行文件

go build

sudo ./collector