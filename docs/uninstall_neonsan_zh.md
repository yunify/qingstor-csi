# 卸载NeonSAN CSI插件

> 注：卸载前，请确保基于 NeonSAN 的 PVC，PV，VolumeSnapshot，VolumeSnapshotContent 已删除。

```
kubectl delete -f ./sts-controller.yaml
kubectl delete -f ./ds-node.yaml
kubectl delete -f ./rbac-controller.yaml
kubectl delete -f ./rbac-node.yaml
kubectl delete -f ./secret-registry.yaml
kubectl delete configmap csi-neonsan -n kube-system
```

