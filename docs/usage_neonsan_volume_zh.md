# NeonSAN CSI 插件用法-存储卷

## 准备
### 创建 StorageClass

StorageClass 是 Kubernetes 的一种资源对象，用来存放存储卷的部分配置。创建此对象是使用 NeonSAN CSI 插件的前提。

```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/sc.yaml
```

- 查询创建的 StorageClass

```
$ kubectl get sc
NAME              PROVISIONER                    AGE
csi-neonsan       csi-neonsan                    1m
```

## 动态创建与删除 PVC


### 创建 PVC


- 创建 PVC
```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc.yaml
persistentvolumeclaim "pvc-test" created
```

- 查询创建的 PVC 与 PV

```
$ kubectl get pvc
NAME       STATUS    VOLUME                 CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-test   Bound     pvc-e281cbd3a9c511e8   10Gi       RWO            csi-neonsan    21s
```

```
$ kubectl get pv
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                  STORAGECLASS   REASON    AGE
pvc-e281cbd3a9c511e8   10Gi       RWO            Delete           Bound       kube-system/pvc-test   csi-neonsan              18s
```

### 删除 PVC

- 删除 PVC
```
$ kubectl delete -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc.yaml
persistentvolumeclaim "pvc-test" deleted
```

- 查询删除的 PVC 和 PV
```
$ kubectl get pvc
No resources found.
```

```
$ kubectl get pv
No resources found.
```

## 静态创建与删除 PVC

### 创建 PVC

- 通过 NeonSAN CLI 创建 image

> 注：请将 NeonSAN CLI 配置文件 qbd.conf 放置在当前文件夹内。
```
$ neonsan create_volume -volume pre-provisioning-volume -pool kube -size 5G -repcount 1 -c ./qbd.conf
INFO[0000] create volume succeed.                       
```

- 创建 PV
```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pv.yaml
persistentvolume "pv-neonsan" created
```

- 查询 PV
```
$ kubectl get pv
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM     STORAGECLASS   REASON    AGE
pv-neonsan             5Gi        RWO            Delete           Available             csi-neonsan              13s
```

- 创建 PVC
```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc.yaml
persistentvolumeclaim "pvc-test" created
```

- 查询 PVC 和 PV
```
$ kubectl get pvc
NAME       STATUS    VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-test   Bound     pv-neonsan   5Gi        RWO            csi-neonsan    3s
```

```
$ kubectl get pv
NAME         CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                    STORAGECLASS   REASON    AGE
pv-neonsan   5Gi        RWO            Delete           Bound       kube-system/pvc-test     csi-neonsan              3m
```

### 删除 PVC

- 删除 PVC
```
$ kubectl delete pvc pvc-test
persistentvolumeclaim "pvc-test" deleted
```

- 删除 PV
```
$ kubectl delete pv pv-neonsan
persistentvolume "pv-neonsan" deleted
```

- 通过 NeonSAN CLI 删除 image
```
$ neonsan delete_volume -pool kube -volume pre-provisioning-volume
delete volume succeed.
```

## 工作负载使用 PVC 持久化数据

### 创建工作负载

- 查询 PVC
```
$ kubectl get pvc
NAME       STATUS    VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-test   Bound     pv-neonsan   5Gi        RWO            csi-neonsan    3m

```

- 创建 Deployment 挂载 PVC

```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/deploy.yaml
deployment.apps "nginx" created
```

- 查询 Pod 状态
```
$ kubectl get po --selector=app=nginx
NAME                    READY     STATUS    RESTARTS   AGE
nginx-7cb56987d-582bx   1/1       Running   0          5m
```

- 查看挂载 PVC 状态
```
$ kubectl exec -ti nginx-7cb56987d-582bx /bin/bash
# ls mnt/
lost+found
```

### 删除工作负载

- 删除 Deployment
```
$ kubectl delete -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/deploy.yaml
deployment.apps "nginx" deleted
```