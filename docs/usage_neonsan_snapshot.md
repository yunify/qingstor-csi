# NeonSAN CSI plugin usage - snapshot

## Prerequsite
### Create PVC

Please reference [NeonSAN CSI plugin usage - volume](./usage_neonsan_volume.md) to create NeonSAN based PVC named `pvc-test`.

### Create snapshot class

```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/snapshot-class.yaml

$ kubectl get volumesnapshotclass
NAME                    AGE
csi-neonsan-snapclass   21h
```

## Usage
### Create snapshot

```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/volume-snapshot.yaml

$ kubectl get volumesnapshot
NAME          AGE
snap-test     21h
```

### Restore volume from snapshot

```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc-restore.yaml

$ kubectl get pvc
NAME            STATUS    VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-restore     Bound     pvc-14c8d0bc-d456-11e8-b49c-5254c1f1ffec   10Gi       RWO            csi-neonsan    21h
pvc-test        Bound     pvc-14c8d0bc-d123-11e8-b49c-5254c1f1ffec   10Gi       RWO            csi-neonsan    21h
```

### Delete snapshot

```
$ kubectl delete -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/volume-snapshot.yaml

$ kubectl get volumesnapshot
No resources found.
```
