# 如何在 Kubernetes 1.12 中安装 NeonSAN CSI 插件

## 准备环境

- NeonSAN 服务端，并创建名为 `kube` 的 Pool 供 Kubernetes 使用
- Kubernetes v1.12+ 集群
    - 启用 [Priviliged Pod](https://kubernetes-csi.github.io/docs/Setup.html#enable-privileged-pods)，将 Kubernetes 的 kubelet 和 kube-apiserver 组件配置项 `--allow-privileged` 设置为 `true`
    - 启用（默认开启）[Mount Propagation](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) 特性
    - 启用 Kubernetes 的 kubelet, controller-manager 和 kube-apiserver 组件的若干 [Feature Gate](https://kubernetes-csi.github.io/docs/Setup.html#enabling-features)
    ```
    --feature-gates=VolumeSnapshotDataSource=true,KubeletPluginsWatcher=true,CSINodeInfo=true,CSIDriverRegistry=true
    ```
- Kubernetes 集群各节点安装 `neonsan` 和 `qbd` NeonSAN CLI 程序并且可以访问 NeonSAN 服务端

## 下载安装包

```
$ wget $(curl --silent "https://api.github.com/repos/yunify/qingstor-csi/releases/latest" | \
  grep browser_download_url | grep install|cut -d '"' -f 4)
$ tar -xvf csi-qingstor-install.tar.gz
$ cd csi-qingstor-install
```
## 设置配置项

### 设置 kubelet 路径

如果Kubernetes集群的Kubelet设置了 `--root-dir` 选项（默认为 *`/var/lib/kubelet`* ），请将 `ds-node.yaml` 文件内 *`/var/lib/kubelet`* 的值替换为 `--root-dir` 选项的值。

### 设置插件启动参数

插件 `protocol` 启动参数设置插件与 NeonSAN 服务端传输文件的协议，可以设置为 `TCP` 或 `RDMA`
修改 `ds-node.yaml` 和 `sts-controller.yaml` 文件 `csi-neonsan` 容器启动参数，使其生效。


## 创建 ConfigMap

- 编辑 NeonSAN CLI 的配置文件 [`qbd.conf`](../deploy/neonsan/kubernetes/qbd.conf)
```
$ vi ./qbd.conf
[zookeeper]
# IP of zookeeper cluster
ip="IP:PORT"
cluster_name="CLUSTER_NAME"

[client]
tcp_no_delay=1
io_depth=64
io_timeout=30
conn_timeout=8
open_volume_timeout=180
```

- 检验 `qbd.conf` 配置文件

```
$ sudo neonsan list_pool -c ./qbd.conf
$ sudo echo $?
0
```

- 创建 ConfigMap 对象
```
$ kubectl create configmap csi-neonsan --from-file=qbd.conf=./qbd.conf --namespace=kube-system
```

## 创建镜像仓库密钥

```
$ kubectl apply -f ./secret-registry.yaml
```

## 创建注册插件对象

```
$ kubectl apply -f ./crd-csidriver.yaml  --validate=false
$ kubectl apply -f ./crd-csinodeinfo.yaml  --validate=false
$ kubectl apply -f ./csidriver.yaml
```

## 创建服务程序

```
$ kubectl apply -f ./rbac-controller.yaml
$ kubectl apply -f ./rbac-node.yaml
$ kubectl apply -f ./sts-controller.yaml
$ kubectl apply -f ./ds-node.yaml
```

## 检查插件状态

```
$ kubectl get pods -n kube-system --selector=app=csi-neonsan
NAME                            READY     STATUS        RESTARTS   AGE
csi-neonsan-controller-0      4/4       Running       0          5m
csi-neonsan-node-kks3q        2/2       Running       0          2m
csi-neonsan-node-pgsbn        2/2       Running       0          2m
```

