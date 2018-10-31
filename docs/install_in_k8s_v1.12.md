# How to install NeonSAN CSI plugin in Kubernetes v1.12

## Prerequsite

- Need a NeonSAN server, and create a `kube` pool used for Kubernetes.
- In Kubernetes v1.12 cluster
    - enable [Priviliged Pod](https://kubernetes-csi.github.io/docs/Setup.html#enable-privileged-pods), set `--allow-privileged` as `true` in Kubernetes `kubelet` and `kube-apiserver` option.
    - enable [Mount Propagation](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) feature gate.
    - enable several [Feature Gate](https://kubernetes-csi.github.io/docs/Setup.html#enabling-features) in Kubernetes `kubelet`, `controller-manager` and `kube-apiserver` option.
    ```
    --feature-gates=VolumeSnapshotDataSource=true,KubeletPluginsWatcher=true,CSINodeInfo=true,CSIDriverRegistry=true
    ```
- In Kubernetes nodes, please install NeonSAN CLI tools, `neonsan` and `qbd`, and ensure that we can connect to NeonSAN server.

## Download

```
$ wget $(curl --silent "https://api.github.com/repos/yunify/qingstor-csi/releases/latest" | \
  grep browser_download_url | grep install|cut -d '"' -f 4)
$ tar -xvf csi-qingstor-install.tar.gz
$ cd csi-qingstor-install
```

## Configuration

### Set kubelet path

> NOTE: If Kubernetes' kubelet already set the `--root-dir` option (default: *"/var/lib/kubelet"*), please replace the value of `spec.template.spec.containers[name=csi-neonsan].volumeMounts[name=mount-dir].mountPath` and `spec.template.spec.volumes[name=mount-dir].hostPath.path` fileds in `csi-node-ds.yaml` file with the value of `--root-dir`.

### Set plugin option

the  `csi-neonsan` container option `protocol` can be set as `TCP` or `RDMA` in `ds-node.yaml` and `sts-controller.yaml` file.

## Create ConfigMap

- Edit NeonSAN Config file [`qbd.conf`](../deploy/neonsan/kubernetes/qbd.conf)
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

## Create Docker image registry secret

```
$ kubectl apply -f ./secret-registry.yaml
```

## Create CRD object

```
$ kubectl apply -f ./crd-csidriver.yaml  --validate=false
$ kubectl apply -f ./crd-csinodeinfo.yaml  --validate=false
$ kubectl apply -f ./csidriver.yaml
```

## Deploy plugin

```
$ kubectl apply -f ./rbac-controller.yaml
$ kubectl apply -f ./rbac-node.yaml
$ kubectl apply -f ./sts-controller.yaml
$ kubectl apply -f ./ds-node.yaml
```

## Check plugin status
```
$ kubectl get pods -n kube-system --selector=app=csi-neonsan
NAME                            READY     STATUS        RESTARTS   AGE
csi-neonsan-controller-0      4/4       Running       0          5m
csi-neonsan-node-kks3q        2/2       Running       0          2m
csi-neonsan-node-pgsbn        2/2       Running       0          2m
```
