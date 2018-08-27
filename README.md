# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunify/qingstor-csi)](https://goreportcard.com/report/github.com/yunify/qingstor-csi)

> English | [中文](README_zh.md)
## Description
QingStor CSI plugin implements an interface between Container Storage Interface ([CSI](https://github.com/container-storage-interface/)) enabled Container Orchestrator (CO) and the storage of QingStor. Currently, QingCloud CSI plugin is tested in Kubernetes v1.11+ environment.

## NeonSAN Storage Plugin

### Prerequisite

User should ensure NeonSAN client command tools, `qbd` and `neonsan`, have been installed on cluster nodes and work normally.  

### Installation

- Edit NeonSAN Config file
```
[zookeeper]
# IP of zookeeper cluster
ip="IP:PORT"
cluster_name="CLUSTER_NAME"

[client]
tcp_no_delay=1
io_depth=64
io_timeout=30
conn_timeout=8
open_volume_timeout=180
```

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

- Create a Deployment

```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/deploy.yaml
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

- `pool`: NeonSAN pool. Default is `kube`.

- `replicas`: number of volume replicas. Default is `1`.

- `stepSize`: volume size increment in GiB. Default is `1`.

- `fsType`: `ext3`, `ext4`, `xfs`. Default `ext4`.

## Support
If you have any qustions or suggestions, please submit an issue at [qingstor-csi](https://github.com/yunify/qingstor-csi/issues).