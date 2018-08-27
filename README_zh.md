# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunify/qingstor-csi)](https://goreportcard.com/report/github.com/yunify/qingstor-csi)

> [English](README.md) | 中文
## 介绍

QingStor CSI 插件实现 CSI 接口，使容器编排平台（如 Kubernetes）能够使用 QingStor 存储产品的资源。目前，已经开发的 NeonSAN CSI插件已经在 Kubernetes v1.11 环境中通过了 CSI 可用性测试。

## 安装

### 准备环境

- NeonSAN 服务端，并创建供 Kubernetes 使用的 Pool
- Kubernetes v1.11+ 集群， 配置项 `--allow-privileged` 设置为 `true`，启用（默认开启）[Mount Propagation](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) 特性
- Kubernetes 集群各节点可使用 NeonSAN CLI 的 `neonsan` 和 `qbd` 访问 NeonSAN 服务端

### 下载安装包

```
$ wget $(curl --silent "https://api.github.com/repos/yunify/qingstor-csi/releases/latest" | \
  grep browser_download_url | grep install|cut -d '"' -f 4)
$ tar -xvf csi-qingstor-install.tar.gz
$ cd csi-qingstor-install
```

### 创建 ConfigMap

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

### 创建 Docker 镜像仓库密钥

```
$ kubectl apply -f ./csi-secret.yaml
```

### 创建访问控制对象

```
$ kubectl apply -f ./csi-controller-rbac.yaml
$ kubectl apply -f ./csi-node-rbac.yaml
```

### 部署 CSI 插件

> 注：如果Kubernetes集群的Kubelet设置了 `--root-dir` 选项（默认为 *"/var/lib/kubelet"*），请将 `csi-node.ds.yaml` 文件 `spec.template.spec.containers[name=csi-neonsan].volumeMounts[name=mount-dir].mountPath` 和 `spec.template.spec.volumes[name=mount-dir].hostPath.path` 的值 *"/var/lib/kubelet"* 替换为 `--root-dir` 选项的值。

```
$ kubectl apply -f ./csi-controller-sts.yaml
$ kubectl apply -f ./csi-node-ds.yaml
```

### 检查 CSI 插件状态

```
$ kubectl get pods -n kube-system --selector=app=csi-neonsan
NAME                            READY     STATUS        RESTARTS   AGE
csi-neonsan-controller-0      3/3       Running       0          5m
csi-neonsan-node-kks3q        2/2       Running       0          2m
csi-neonsan-node-pgsbn        2/2       Running       0          2m
```


## 参考资料

- [版本依赖](docs/reference_zh.md)
- [版本特性](docs/reference_zh.md)
- [StorageClass 参数说明](docs/reference_zh.md)
- [卸载 NeonSAN CSI 插件](docs/uninstall_neonsan_zh.md)
- [使用 NeonSAN CSI 插件](docs/usage_neonsan_zh.md)

## 支持
如果有任何问题或建议, 请在 [qingstor-csi](https://github.com/yunify/qingstor-csi/issues) 项目提 issue。
