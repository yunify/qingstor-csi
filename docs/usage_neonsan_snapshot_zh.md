# NeonSAN CSI 插件用法-快照

## 准备
### 创建 PVC

参照 [NeonSAN CSI 插件用法-存储卷](docs/usage_neonsan_volume_zh.md) 创建基于 NeonSAN 的 PVC，PVC 名为 pvc-test

### 创建快照类型

```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/snapshot-class.yaml

$ kubectl get volumesnapshotclass
NAME                    AGE
csi-neonsan-snapclass   21h
```

## 功能
### 创建快照

```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/volume-snapshot.yaml

$ kubectl get volumesnapshot
NAME          AGE
snap-test     21h
```

### 从快照创建存储卷

```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc-restore.yaml

$ kubectl get pvc
NAME            STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-restore     Bound     pvc-14c8d0bc-d456-11e8-b49c-5254c1f1ffec   10Gi       RWO            csi-neonsan    21h
pvc-test        Bound     pvc-14c8d0bc-d123-11e8-b49c-5254c1f1ffec   10Gi       RWO            csi-neonsan    21h
```

### 删除快照

```
$ kubectl delete -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/volume-snapshot.yaml

$ kubectl get volumesnapshot
No resources found.
```
