
# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunify/qingstor-csi)](https://goreportcard.com/report/github.com/yunify/qingstor-csi)

## Description
QingStor CSI plugin implements an interface between Container Storage Interface ([CSI](https://github.com/container-storage-interface/)) enabled Container Orchestrator (CO) and the storage of NeonSAN, which has passed [CSI sanity test](https://github.com/kubernetes-csi/csi-test). 

## Installation
This guide will install CSI plugin in the *kube-system* namespace of Kubernetes v1.14+. You can also deploy the plugin in other namespace. 

- Set Kubernetes Parameters
  - Enable `--allow-privileged=true` on kube-apiserver, kube-controller-manager, kube-scheduler, kubelet
  - Enable (Default enabled) [Mount Propagation](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) feature gate。
  - Enable `--feature-gates=CSINodeInfo=true,CSIDriverRegistry=true,KubeletPluginsWatcher=true` option on kube-apiserver, kube-controller-manager, kube-scheduler, kubelet
  
- Download **qbd** and install **qbd** on nodes of kubernetes
  As long as **qbd**'s version consistent with neonsan server, the CSI works.

  * Download
  
    As **qbd** is not open source,  the install package is provided by **Neonsan Team**
  
  * Install
  
  | OS            | Required lib            | Command                            |
  | :------------ | :---------------------- | :--------------------------------- |
  | Redhat/Centos | libcurl libicu          | rpm -ivh pitrix-dep-qbd-xxx.rpm    |
  | SUSE          | libcurl4 libicu         | rpm -ivh pitrix-dep-qbd-xxx.rpm    |
  | Ubuntu/Debian | libcurl4-openssl libicu | apt install pitrix-dep-qbd-xxx.deb |
  
  * Check installed
  
    ```
    $ qbd -v
    Package Version:       2.0.4-cb3daa5-190821224030-centos75
    Loaded Module Version: 2.0.4-cb3daa5-190821224030-centos75
    NeonSAN Static Library Version: 2.1.14-83d762a`
    ```


- Deploy CSI plugin
  ```
  kubectl apply -f deploy/neonsan/kubernetes/release/csi-neonsan-v1.2.0.yaml
  ```

- Check CSI plugin
  ```
  kubectl get pods -n kube-system --selector=app=csi-neonsan
  NAME                                     READY   STATUS    RESTARTS   AGE
  csi-neonsan-controller-594448465-sq57l   4/4     Running   0          6m41s
  csi-neonsan-node-9w2zp                   1/1     Running   0          6m41s
  csi-neonsan-node-bzqcj                   1/1     Running   0          6m41s
  csi-neonsan-node-vjmvb                   1/1     Running   0          6m41s
  ```

- Install neonsan-plugin
  ```
  make neonsan-plugin
  ansible-playbook deploy/neonsan/plugin/neonsan-plugin-install.yaml
  ``` 
  
- Check neonsan-plugin
   ``` 
  ansible all -m shell -a "systemctl status neonsan-plugin.service" | grep active
     Active: active (running) since Tue 2020-02-25 10:42:25 CST; 40s ago
     Active: active (running) since Tue 2020-02-25 10:42:25 CST; 40s ago
     Active: active (running) since Tue 2020-02-25 10:42:25 CST; 40s ago
   ``` 

### Uninstall

``` 
  ansible-playbook deploy/neonsan/plugin/neonsan-plugin-uninstall.yaml
  kubectl delete -f deploy/neonsan/kubernetes/release/csi-neonsan-v1.2.0.yaml
``` 

### StorageClass Parameters
StorageClass definition [file](deploy/neonsan/example/volume/sc.yaml) shown below is used to create StorageClass object.

```
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: csi-neonsan
  provisioner: neonsan.csi.qingcloud.com
  parameters:
    fsType: "ext4"
    replica: "2"
    pool: "kube"
  reclaimPolicy: Delete 
```

- `fsType`: `ext3`, `ext4`, `xfs`. Default `ext4`.
- `replica`: count of replicas (`1-3`). Default` 1`.
- `poolName`: pool of Neonsan, should not be empty. 
## Support
If you have any questions or suggestions, please submit an issue at [qingstor-csi](https://github.com/yunify/qingstor-csi/issues).
