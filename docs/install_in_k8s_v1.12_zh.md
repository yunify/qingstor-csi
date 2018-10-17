# 如何在 Kubernetes 1.12 中安装 NeonSAN CSI 插件

## 准备环境

- NeonSAN 服务端，并创建供 Kubernetes 使用的 Pool
- Kubernetes v1.12+ 集群
    - 启用 [Priviliged Pod](https://kubernetes-csi.github.io/docs/Setup.html#enable-privileged-pods)，将配置项 `--allow-privileged` 设置为 `true`
    - 启用（默认开启）[Mount Propagation](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) 特性
    - 启用若干 [Feature Gate](https://kubernetes-csi.github.io/docs/Setup.html#enabling-features)
- Kubernetes 集群各节点安装 `neonsan` 和 `qbd` NeonSAN CLI 程序并且可以访问 NeonSAN 服务端

## 下载安装包

```
$ wget $(curl --silent "https://api.github.com/repos/yunify/qingstor-csi/releases/latest" | \
  grep browser_download_url | grep install|cut -d '"' -f 4)
$ tar -xvf csi-qingstor-install.tar.gz
$ cd csi-qingstor-install
```

## 创建 ConfigMap

- 编辑 NeonSAN CLI 的配置文件 [`qbd.conf`](./deploy/neonsan/kubernetes/qbd.conf)
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
$ kubectl apply -f ./csi-secret.yaml
```

## 创建快照相关对象

```
$ kubectl apply -f ./crd-csidriver.yaml
$ kubectl apply -f ./crd-csinodeinfo.yaml
$ kubectl apply -f ./csidriver-neonsan.yaml
$ kubectl apply -f ./crd-snapshotclass.yaml
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

