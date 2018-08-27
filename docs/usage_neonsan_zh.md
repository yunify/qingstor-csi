# 使用 NeonSAN CSI 插件

- 创建 StorageClass
- 动态创建 & 删除 PVC
- 静态创建 & 删除 PVC
- 工作负载使用 PVC 持久化数据

## 创建 StorageClass

StorageClass 是 Kubernetes 的一种资源对象，用来存放存储卷的部分配置。创建此对象是使用 NeonSAN CSI 插件的前提。

```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/sc.yaml
```

- 查询创建的 StorageClass

```
$ kubectl get sc
NAME              PROVISIONER                    AGE
csi-neonsan       csi-neonsan                    1m
```

## 动态创建 & 删除 PVC

- 动态创建 PVC

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

## 静态创建 & 删除 PVC

- 通过 NeonSAN CLI 创建 image
```
$ neonsan create_volume -volume pre-provisioning-volume -pool csi -size 5G -repcount 1
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
