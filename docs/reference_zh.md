# 参考资料

## 版本依赖

| 存储类型| 插件版本 | 分支| CSI 版本 | Kubernetes 版本 | CSI 可用性测试 | NeonSAN 版本|
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
|NeonSAN|v0.3.0-alpha.1|Master| v0.3.0| v1.11.0|v0.3.0-1 | client: dev  |

## 版本特性

| 插件版本  | 创建和删除存储卷  | 挂载和卸载存储卷至工作负载 | 创建和删除快照 | 从快照创建存储卷 | CSI 可用性测试 |
|:---:|:---:|:-------------------------:|:----------:|:---------:|:--------:|:-------:|
| v0.3.0-alpha.1|✓|✓|-|-|✓|

> - `✓`: 支持
> - `-`: 不支持

## NeonSAN

### StorageClass 参数

如下所示的 StorageClass 资源定义文件可用来创建 StorageClass 对象.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: csi-neonsan
provisioner: csi-neonsan
parameters:
  pool: "kube"
  replicas: "1"
  stepSize: "10"
  fsType: "ext4"
reclaimPolicy: Delete 
```

- `pool`: Kubernetes 插件从哪个 pool 内创建存储卷。默认值为 `kube`
- `replicas`: NeonSAN image 的副本个数，默认为 `1`
- `stepSize`: 用户所创建存储卷容量的增量，单位为GiB，默认为 `1`
- `fsType`: 存储卷的文件系统格式，默认为 `ext4`
