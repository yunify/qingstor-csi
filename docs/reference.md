# Reference

## Dependency

| Storage Type| Plugin Version | Branch| CSI Version | Kubernetes Version | Sanity Test | NeonSAN Version|
|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
|NeonSAN|v0.3.0-alpha.1|Master| v0.3.0| v1.11.0|v0.3.0-1 | client: dev  |

## Feature

| Plugin Version   | Create & Delete Volume  | Mount & Unmount Volume on Workload | Create & Delete Snapshot | Create Volume from Snapshot | Sanity Test |
|:---:|:---:|:-------------------------:|:----------:|:---------:|:--------:|:-------:|
| v0.3.0-alpha.1|✓|✓|-|-|✓|

> - `✓`: Supported
> - `-`: Unsupported

## NeonSAN

### StorageClass Parameters

StorageClass definition file shown below is used to create StorageClass object.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: csi-neonsan
provisioner: csi-neonsan
parameters:
  pool: "kube"
  replicas: "1"
  stepSize: "10"
  fsType: "ext4"
reclaimPolicy: Delete 
```

- `pool`: the name of the volume storage pool. Default `kube`.
- `replicas`: NeonSAN image replica count. Default: `1`.
- `stepSize`: set the increment of volumes size in GiB.，Default: `1`.
- `fsType`: the file system to use for the volume. Default: `ext4`.