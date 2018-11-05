# 如何卸载NeonSAN CSI插件


> 注：卸载前，请确保基于 NeonSAN 的 PVC，PV，VolumeSnapshot，VolumeSnapshotContent 已删除。

在安装文件夹内执行下列命令：
```
kubectl delete -f ./sts-controller.yaml
kubectl delete -f ./ds-node.yaml
kubectl delete -f ./rbac-controller.yaml
kubectl delete -f ./rbac-node.yaml
kubectl delete -f ./secret-registry.yaml
kubectl delete -f ./csidriver.yaml
kubectl delete configmap csi-neonsan -n kube-system
```