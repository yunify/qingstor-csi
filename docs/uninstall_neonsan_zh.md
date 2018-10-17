# 卸载NeonSAN CSI插件

> 注：卸载前，请确保基于 NeonSAN 的 PVC，PV，VolumeSnapshot，VolumeSnapshotContent 已删除。

```
$ kubectl delete -f ./csi-controller-sts.yaml
$ kubectl delete -f ./csi-node-ds.yaml
$ kubectl delete -f ./csi-controller-rbac.yaml
$ kubectl delete -f ./csi-node-rbac.yaml
$ kubectl delete -f ./csi-secret.yaml
$ kubectl delete configmap csi-neonsan -n kube-system
```

