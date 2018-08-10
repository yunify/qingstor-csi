# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)

## Description
TBD

## NeonSAN Storage Plugin

### Installation
- Install NeonSAN client tool on Kubernetes masters and nodes.

- Create ConfigMap
```
$ kubectl create configmap csi-neonsan --from-file=qbd.conf=./qbd.conf --namespace=kube-system
```

- Create Docker image registry secret
```
$ kubectl apply -f ./csi-secret.yaml
```

- Create access control object
```
$ kubectl apply -f ./csi-controller-rbac.yaml
$ kubectl apply -f ./csi-node-rbac.yaml
```

- Deploy CSI plugin
```
$ kubectl apply -f ./csi-controller-sts.yaml
$ kubectl apply -f ./csi-node-ds.yaml
```

### Simple Example

- Create a StorageClass

```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/sc.yaml
```

- Create a PVC

```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc.yaml
```

### StorageClass Parameters
StorageClass definition [file](deploy/neonsan/example/sc.yaml) shown below is used to create StorageClass object.
```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: csi-neonsan
provisioner: csi-neonsan
parameters:
  pool: "csi"
  replicas: "1"
  stepSize: "10"
  fsType: "ext4"
reclaimPolicy: Delete
```

- `pool`: NeonSAN pool. Default is `csi`.

- `replicas`: number of volume replicas. Default is `1`.

- `stepSize`: volume size increment in GiB. Default is `1`.

- `fsType`: `ext3`, `ext4`, `xfs`. Default `ext4`.


## Support
If you have any qustions or suggestions, please submit an issue at [qingstor-csi](https://github.com/yunify/qingstor-csi/issues)
