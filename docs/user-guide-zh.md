# 使用指南

## 如何设置存储类型
### 存储类型模版

如下所示的 StorageClass 资源定义可用来创建 StorageClass 对象。
```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: csi-neonsan
provisioner: neonsan.csi.qingcloud.com
parameters:
  fsType: "ext4"
  replica: "1"
reclaimPolicy: Delete
allowVolumeExpansion: true
```

### 存储卷参数
存储卷类型模板中 `.parameters` 设置存储卷参数

#### `fsType`
支持 `ext3`, `ext4`, `xfs`. 默认为 `ext4`。

#### `replica`
代表单副本硬盘副本数。 默认为 `1`，最大为`3`。

### 其他参数

#### 设置默认存储类型
存储类型模版中 `.metadata.annotations.storageclass.beta.kubernetes.io/is-default-class` 的值设置为 `true` 表明此存储类型设置为默认存储类型。详见 [Kubernetes 官方文档](https://kubernetes.io/docs/tasks/administer-cluster/change-default-storage-class/)

#### 扩容
存储类型模版中 `.allowVolumeExpansion` 的值可填写 `true` 或 `false`, 设置是否支持扩容存储卷。详见 [Kubernetes 官方文档](https://kubernetes.io/docs/concepts/storage/storage-classes/#allow-volume-expansion)

## 存储卷管理
存储卷（PVC，PersistentVolumeClaim）管理功能包括动态分配存储卷，删除存储卷，挂载存储卷到 Pod，从 Pod 卸载存储卷。

### 准备工作
- Kubernetes 1.14+ 集群
- 安装了 QingStor CSI 存储插件
- 安装了 QingStor CSI 存储类型

#### 安装 QingCloud CSI 存储类型
- 安装
```console
$ kubectl create -f sc.yaml
```
- 检查
```console
$ kubectl get sc
NAME            PROVISIONER              AGE
csi-neonsan neonsan.csi.qingcloud.com   14m
```

### 创建存储卷
- 创建存储卷
```console
$ kubectl create -f pvc.yaml 
persistentvolumeclaim/pvc-test created
```
- 检查存储卷
```console
$ kubectl get pvc pvc-test
NAME          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
pvc-test      Bound    pvc-76429525-a930-11e9-9a6a-5254ef68c8c1   20Gi       RWO            csi-neonsan     25m
```

### 挂载存储卷
- 创建 Deployment 挂载存储卷
```console
$ kubectl create -f deploy-nginx.yaml 
deployment.apps/nginx created
```
- 访问容器内挂载存储卷的目录
```console
$ kubectl exec -ti nginx-84474cf674-zfhbs /bin/bash
# cd /mnt
# ls
lost+found
```

### 卸载存储卷
- 删除 deployment 卸载存储卷
```console
$ kubectl delete deploy nginx
deployment.extensions "nginx" deleted
```

### 删除存储卷
- 删除存储卷
```console
$ kubectl delete pvc pvc-test
persistentvolumeclaim "pvc-test" deleted
```
- 检查存储卷
```console
$ kubectl get pvc pvc-test
Error from server (NotFound): persistentvolumeclaims "pvc-example" not found
```

## 存储卷扩容
扩容功能将扩大存储卷可用容量。由于平台限制，本存储插件仅支持离线扩容硬盘。离线扩容硬盘流程是 1. 存储卷处于未挂载状态，2. 扩容存储卷，3. 挂载一次存储卷。示例 YAML 文件在 https://github.com/yunify/qingcloud-csi/tree/master/deploy/disk/volume 内。

### 准备工作
- Kubernetes 1.14+ 集群
- Kubernetes 组件的 `feature-gate` 增加 `ExpandCSIVolumes=true`
- 配置 QingCloud CSI 存储类型，并将其 `allowVolumeExpansion` 字段值设置为 `true`
- 创建一个存储卷并挂载至 Pod，参考存储卷管理

### 卸载存储卷
```console
$ kubectl scale deploy nginx --replicas=0
```

### 扩容存储卷
- 修改存储卷容量
```console
$ kubectl patch pvc pvc-test -p '{"spec":{"resources":{"requests":{"storage": "40Gi"}}}}'
persistentvolumeclaim/pvc-test patched
```
- 挂载存储卷
```console
$ kubectl scale deploy nginx --replicas=1
```
- 完成扩容
```console
$ kubectl get pvc pvc-test
NAME          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
pvc-test   Bound    pvc-906f5760-a935-11e9-9a6a-5254ef68c8c1   40Gi       RWO            csi-qingcloud   6m7s
$ kubectl get pod
NAME                     READY   STATUS    RESTARTS   AGE
nginx-6c444c9b7f-d6n29   1/1     Running   0          3m38s
```

### 检查
- 进入 Pod 查看
```console
$ kubectl exec -ti nginx-6c444c9b7f-d6n29 /bin/bash
root@nginx-6c444c9b7f-d6n29:/# df -ah
Filesystem      Size  Used Avail Use% Mounted on
...
/dev/vdc         40G   49M   40G   1% /mnt
...
```

