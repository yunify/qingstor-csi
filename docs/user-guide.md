# User Guide 

## Set Storage Class 
### An example of Storage Class 

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: csi-neonsan
provisioner: neonsan.csi.qingstor.com
parameters:
  fsType: "ext4"
  replica: "1"
  pool: "kube"
reclaimPolicy: Delete
allowVolumeExpansion: true
```

### Parameters in Storage Class

#### `fsType`
Support `ext3`, `ext4`, `xfs`, default `ext4`.

#### `replica`
Number of disk replicas, default `1`，maximum `3`.

#### `pool`
Neonsan pool，not empty

### Other Parameters 

#### Set Default Storage Class
Set the annotation `.metadata.annotations.storageclass.beta.kubernetes.io/is-default-class` value as `true`. See details in [Kubernetes docs](https://kubernetes.io/docs/tasks/administer-cluster/change-default-storage-class/)

#### Volume Expansion
Set the value of `.allowVolumeExpansion` as `true`. See details in [Kubernetes docs](https://kubernetes.io/docs/concepts/storage/storage-classes/#allow-volume-expansion)

## Volume Management 
Volume management including dynamical provisioning/deleting volume, attaching/detaching volume. Please reference [Example YAML Files](https://github.com/yunify/qingcloud-csi/tree/master/deploy/disk/example/volume).

### Prerequisite
- Kubernetes 1.14+ 
- Neonsan CSI installed 
- Neonsan CSI StorageClass created 

#### Create Storage Class 
- Create 
```console
$ kubectl create -f sc.yaml
```
- Check 
```console
$ kubectl get sc
NAME            PROVISIONER              AGE
csi-neonsan neonsan.csi.qingstor.com   14m
```

### Create PVC 
- Create
```console
$ kubectl create -f pvc.yaml 
persistentvolumeclaim/pvc-test created
```
- Check 
```console
$ kubectl get pvc pvc-test
NAME          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
pvc-test      Bound    pvc-76429525-a930-11e9-9a6a-5254ef68c8c1   20Gi       RWO            csi-neonsan     25m
```

### Mount Volume
- Create Deployment 
```console
$ kubectl create -f deploy-nginx.yaml 
deployment.apps/nginx created
```
- Check 
```console
$ kubectl exec -ti nginx-84474cf674-zfhbs /bin/bash
# cd /mnt
# ls
lost+found
```

### Unmount Volume
- Delete Deployment 
```console
$ kubectl delete deploy nginx
deployment.extensions "nginx" deleted
```

### Delete PVC 
- Delete 
```console
$ kubectl delete pvc pvc-test
persistentvolumeclaim "pvc-test" deleted
```
- Check 
```console
$ kubectl get pvc pvc-test
Error from server (NotFound): persistentvolumeclaims "pvc-example" not found
```

## Volume Expansion 
This plugin only supports offline volume expansion. The procedure of offline volume expansion is shown as follows. 
1. Ensure volume in unmounted status
2. Edit the capacity of PVC
3. Mount volume on workload
Please reference [Example YAML files](https://github.com/yunify/qingcloud-csi/tree/master/deploy/disk/volume).

### Prerequisite
- Kubernetes 1.14+ cluster
- Add `ExpandCSIVolumes=true` in `feature-gate` 
- Set `allowVolumeExpansion` as `true` in StorageClass
- Create a Pod mounting a volume

### Unmount Volume 
```console
$ kubectl scale deploy nginx --replicas=0
```

### Expand Volume
- Change Volume capacity 
```console
$ kubectl patch pvc pvc-test -p '{"spec":{"resources":{"requests":{"storage": "40Gi"}}}}'
persistentvolumeclaim/pvc-test patched
```
- Mount Volume 
```console
$ kubectl scale deploy nginx --replicas=1
```
- Check PVC Capacity 
```console
$ kubectl get pvc pvc-test
NAME          STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
pvc-test   Bound    pvc-906f5760-a935-11e9-9a6a-5254ef68c8c1   40Gi       RWO            csi-qingcloud   6m7s
$ kubectl get pod
NAME                     READY   STATUS    RESTARTS   AGE
nginx-6c444c9b7f-d6n29   1/1     Running   0          3m38s
```

### Check 
```console
$ kubectl exec -ti nginx-6c444c9b7f-d6n29 /bin/bash
root@nginx-6c444c9b7f-d6n29:/# df -ah
Filesystem      Size  Used Avail Use% Mounted on
...
/dev/vdc         40G   49M   40G   1% /mnt
...
```

## Volume Cloning 
Cloning is defined as a duplicate of an existing PVC. Please reference [Example YAML](deploy/neonsan/example/volume) 

### Prerequisites
- Kubernetes 1.15+  
- Enable `VolumePVCDataSource=true` feature gat
- Neonsan CSI installed
- Neonsan CSI StorageClass created 
- Source PVC created 

### Clone Volume 
- Check source PVC 

```console
$ kubectl get pvc pvc-test
NAME       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-test   Bound    pvc-3bdbde24-7016-430e-b217-9eca185caca3   20Gi       RWO            csi-neonsan    3h16
```

- Clone Volume 
```console
$ kubectl create -f pvc-clone.yaml
persistentvolumeclaim/pvc-clone created
```

- Check 
```console
$ kubectl get pvc pvc-clone
NAME        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-clone   Bound    pvc-a75e3f7c-59af-43ef-82d3-300508871432   20Gi       RWO            csi-neonsan    7m4s
```

## Snapshot Management
Snapshot management contains creating/deleting snapshot and restoring volume from snapshpot. Please reference [Example YAML files](https://github.com/yunify/qingcloud-csi/tree/master/deploy/disk/example/snapshot).

### Prerequisites
- Kubernetes 1.14+ 
- Enable `VolumeSnapshotDataSource=true` feature gate at kube-apiserver and kube-controller-manager
- Neonsan CSI v1.2.0 installed
- Neonsan CSI StorageClass created
- Source PVC created 

#### Create Source PVC `pvc-source`
- Create 
```console
$ kubectl create -f pvc-source.yaml
persistentvolumeclaim/pvc-source created
```
- Check 
```console
$ kubectl get pvc pvc-source
NAME         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
pvc-source   Bound    pvc-3bdbde24-7016-430e-b217-9eca185caca3   20Gi       RWO            csi-neonsan    4h25m
```

### Create Volume Snapshot 

#### Create VolumeSnapshotClass 
```console
$ kubectl create -f snapshot-class.yaml
volumesnapshotclass.snapshot.storage.k8s.io/csi-neonsan created
 
$ kubectl get volumesnapshotclass
NAME            AGE
csi-neonsdan    16s
```

#### Create VolumeSnapshot 
```console
$ kubectl create -f snapshot.yaml 
volumesnapshot.snapshot.storage.k8s.io/snap-1 created

$ kubectl get volumesnapshot
NAME     AGE
snap-1   91s
```

### Restore from VolumeSnapshot 
####  create PVC `pvc-snap` from snapshot
```console
$ kubectl create -f pvc-snapshot.yaml 
persistentvolumeclaim/pvc-snap created
```

```console
$ kubectl get pvc pvc-snap
NAME       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-snap   Bound    pvc-a56f6ebe-b37b-40d7-bfb7-aafbecb6672b   20Gi       RWO            csi-neonsan    59m
```

### Delete VolumeSnapshot 
```console
$ kubectl delete volumesnapshot snap-1
volumesnapshot.snapshot.storage.k8s.io "snap-1" deleted
```
