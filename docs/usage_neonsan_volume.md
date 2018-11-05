# NeonSAN CSI plugin usage - volume

## Prerequsite
### Create StorageClass

StorageClass is a kind of Kubernetes resource object to store a part of volume setting. Before using NeonSAN CSI plugin, please create StorageClass. 

```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/sc.yaml
```

- Find StorageClass

```
$ kubectl get sc
NAME              PROVISIONER                    AGE
csi-neonsan       csi-neonsan                    1m
```

## Dynamic volume provisioning


### Create PVC


- Create PVC
```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc.yaml
persistentvolumeclaim "pvc-test" created
```

- Find PVC and PV

```
$ kubectl get pvc
NAME       STATUS    VOLUME                 CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-test   Bound     pvc-e281cbd3a9c511e8   10Gi       RWO            csi-neonsan    21s
```

```
$ kubectl get pv
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                  STORAGECLASS   REASON    AGE
pvc-e281cbd3a9c511e8   10Gi       RWO            Delete           Bound       kube-system/pvc-test   csi-neonsan              18s
```

### Delete PVC

- Delete PVC
```
$ kubectl delete -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc.yaml
persistentvolumeclaim "pvc-test" deleted
```

- Find PVC and PV
```
$ kubectl get pvc
No resources found.
```

```
$ kubectl get pv
No resources found.
```

## Static volume provisioning

### Create PVC

-  Create NeonSAN volume

> IMPORTANT: Please ensure `qbd.conf` file, the config file of NeonSAN CLI, is in the current directory.
```
$ neonsan create_volume -volume pre-provisioning-volume -pool kube -size 5G -repcount 1 -c ./qbd.conf
INFO[0000] create volume succeed.                       
```

- Create PV
```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pv.yaml
persistentvolume "pv-neonsan" created
```

- Find PV
```
$ kubectl get pv
NAME                   CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM     STORAGECLASS   REASON    AGE
pv-neonsan             5Gi        RWO            Delete           Available             csi-neonsan              13s
```

- Create PVC
```
$ kubectl create -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/pvc.yaml
persistentvolumeclaim "pvc-test" created
```

- Find PVC and PV
```
$ kubectl get pvc
NAME       STATUS    VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-test   Bound     pv-neonsan   5Gi        RWO            csi-neonsan    3s
```

```
$ kubectl get pv
NAME         CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS      CLAIM                    STORAGECLASS   REASON    AGE
pv-neonsan   5Gi        RWO            Delete           Bound       kube-system/pvc-test     csi-neonsan              3m
```

### Delete PVC

- Delete PVC
```
$ kubectl delete pvc pvc-test
persistentvolumeclaim "pvc-test" deleted
```

- Delete PV
```
$ kubectl delete pv pv-neonsan
persistentvolume "pv-neonsan" deleted
```

- Delete NeonSAN volume
```
$ neonsan delete_volume -pool kube -volume pre-provisioning-volume
delete volume succeed.
```

## Mount PVC on workload

### Create workload

- Find PVC
```
$ kubectl get pvc
NAME       STATUS    VOLUME       CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-test   Bound     pv-neonsan   5Gi        RWO            csi-neonsan    3m

```

- Create Deployment mounting PVC

```
$ kubectl apply -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/deploy.yaml
deployment.apps "nginx" created
```

- Find Pod
```
$ kubectl get po --selector=app=nginx
NAME                    READY     STATUS    RESTARTS   AGE
nginx-7cb56987d-582bx   1/1       Running   0          5m
```

- Check PVC in workload
```
$ kubectl exec -ti nginx-7cb56987d-582bx /bin/bash
# ls mnt/
lost+found
```

### Delete workload

- Delete Deployment
```
$ kubectl delete -f https://raw.githubusercontent.com/yunify/qingstor-csi/master/deploy/neonsan/example/deploy.yaml
deployment.apps "nginx" deleted
```