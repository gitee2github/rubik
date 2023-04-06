# 约束限制

## 规格
- 磁盘：1GB+ 
- 内存：100MB
- 单个节点Pod上限：100个
    
## 运行时
- 每个kubernetes节点只能部署一个rubik，多个rubik会冲突。
- 除`-v`外，rubik不接受任何命令行参数，否则无法启动。
- 如果rubik进程进入T、D状态，则服务端不可用，此时服务不会响应任何请求。为了避免此情况的发生，请在客户端设置超时时间，避免无限等待。