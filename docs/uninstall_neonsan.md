# How to uninstall NeonSAN CSI plugin

> IMPORTANT: Before uninstalling, please ensure NeonSAN based objects, such as PVC, PV, VolumeSnapshot and VolumeSnapshotContent, have been deleted.

Please execute following commands in the installation package.
```
kubectl delete -f ./sts-controller.yaml
kubectl delete -f ./ds-node.yaml
kubectl delete -f ./rbac-controller.yaml
kubectl delete -f ./rbac-node.yaml
kubectl delete -f ./secret-registry.yaml
kubectl delete -f ./csidriver.yaml
kubectl delete configmap csi-neonsan -n kube-system
```