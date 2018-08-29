# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunify/qingstor-csi)](https://goreportcard.com/report/github.com/yunify/qingstor-csi)

> English | [中文](README_zh.md)
## Description
QingStor CSI plugin implements an interface between Container Storage Interface ([CSI](https://github.com/container-storage-interface/)) enabled Container Orchestrator (CO) and the storage of QingStor. Currently, QingCloud CSI plugin is tested in Kubernetes v1.11+ environment.

## Installation

### Prerequisite
 
- Need a NeonSAN server with a pool used for Kubernetes.
- In Kubernetes v1.11+ cluster, set `--allow-privileged` as `true` and enable [`Mount Propagation`](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) feature gate.
- In Kubernetes cluster, we can manipulate NeonSAN server side through NeonSAN CLI tools, `neonsan` and `qbd`.


### Download
```
$ wget $(curl --silent "https://api.github.com/repos/yunify/qingstor-csi/releases/latest" | \
  grep browser_download_url | grep install|cut -d '"' -f 4)
$ tar -xvf csi-qingstor-install.tar.gz
$ cd csi-qingstor-install
```

### Installation

- Edit NeonSAN Config file [`qbd.conf`](./deploy/neonsan/kubernetes/qbd.conf)
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

- Verify `qbd.conf`

```
$ sudo neonsan list_pool -c ./qbd.conf
$ sudo echo $?
0
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

> NOTE: If Kubernetes' kubelet already set the `--root-dir` option (default: *"/var/lib/kubelet"*), please replace the value of `spec.template.spec.containers[name=csi-neonsan].volumeMounts[name=mount-dir].mountPath` and `spec.template.spec.volumes[name=mount-dir].hostPath.path` fileds in `csi-node-ds.yaml` file with the value of `--root-dir`.
```
$ kubectl apply -f ./csi-controller-sts.yaml
$ kubectl apply -f ./csi-node-ds.yaml
```

- Check CSI plugin status
```
$ kubectl get pods -n kube-system --selector=app=csi-neonsan
NAME                            READY     STATUS        RESTARTS   AGE
csi-neonsan-controller-0      3/3       Running       0          5m
csi-neonsan-node-kks3q        2/2       Running       0          2m
csi-neonsan-node-pgsbn        2/2       Running       0          2m
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

- `fsType`: the file system to use for the volume. Default `ext4`.

## Support
If you have any qustions or suggestions, please submit an issue at [qingstor-csi](https://github.com/yunify/qingstor-csi/issues).