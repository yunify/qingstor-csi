
# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunify/qingstor-csi)](https://goreportcard.com/report/github.com/yunify/qingstor-csi)

## Description
QingStor CSI plugin implements an interface between Container Storage Interface ([CSI](https://github.com/container-storage-interface/)) enabled Container Orchestrator (CO) and the storage of NeonSAN, which has passed [CSI sanity test](https://github.com/kubernetes-csi/csi-test). 

## Notes
- On Kubernetes v1.16, QingStor CSI v1.2.0 not supports volume snapshot management.

## Installation
This guide will install CSI plugin in the *kube-system* namespace of Kubernetes v1.14+. You can also deploy the plugin in other namespace. 

- Set Kubernetes Parameters
  - Enable `--allow-privileged=true` on kube-apiserver, kube-controller-manager, kube-scheduler, kubelet
  - Enable (Default enabled) [Mount Propagation](https://kubernetes.io/docs/concepts/storage/volumes/#mount-propagation) feature gate。
  - Enable `--feature-gates=CSINodeInfo=true,CSIDriverRegistry=true,KubeletPluginsWatcher=true` option on kube-apiserver, kube-controller-manager, kube-scheduler, kubelet

- Install Snapshot CRD and Controller
  
  Snapshot management supported on 1.17+, and CRD and controller should be installed.
   ```bash
   kubectl apply -f deploy/neonsan/kubernetes/snapshot/snapshot-crd.yaml
   kubectl apply -f deploy/neonsan/kubernetes/snapshot/snapshot-controller.yaml
  ``` 

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
  
    ```bash
    qbd -v
    Package Version:       2.0.4-cb3daa5-190821224030-centos75
    Loaded Module Version: 2.0.4-cb3daa5-190821224030-centos75
    NeonSAN Static Library Version: 2.1.14-83d762a
    ```

- Deploy CSI plugin
  - Helm  
    - Install
     ```bash
      helm install ./helm/csi-neonsan --name-template csi-neonsan --namespace kube-system
    ```
    - Check
    ```bash
       helm ls --namespace kube-system
       NAME            NAMESPACE       REVISION        UPDATED                                 STATUS          CHART                           APP VERSION
       csi-neonsan     kube-system     1               2020-05-15 16:15:32.866234841 +0800 CST deployed        csi-neonsan-1.2.0-canary        1.2.0
       kubectl get pods -n kube-system --selector=app=csi-neonsan 
       NAME                                     READY   STATUS    RESTARTS   AGE
       csi-neonsan-controller-594448465-sq57l   4/4     Running   0          6m41s
       csi-neonsan-node-9w2zp                   1/1     Running   0          6m41s
       csi-neonsan-node-bzqcj                   1/1     Running   0          6m41s
       csi-neonsan-node-vjmvb                   1/1     Running   0          6m41s
    ```
   
  - Kubectl
    - For kubernetes 1.16
      ```bash
      kubectl apply -f deploy/neonsan/kubernetes/release/csi-neonsan-v1.2.0-k8s16.yaml
      ```
    - For kubernetes 1.17
      ```bash
      kubectl apply -f deploy/neonsan/kubernetes/release/csi-neonsan-v1.2.0.yaml
      ```
    - Check CSI plugin
      ```bash
      kubectl get pods -n kube-system --selector=app=csi-neonsan
      NAME                                     READY   STATUS    RESTARTS   AGE
      csi-neonsan-controller-594448465-sq57l   4/4     Running   0          6m41s
      csi-neonsan-node-9w2zp                   1/1     Running   0          6m41s
      csi-neonsan-node-bzqcj                   1/1     Running   0          6m41s
      csi-neonsan-node-vjmvb                   1/1     Running   0          6m41s
      ```

- Install neonsan-plugin
    
   As the current version of **qbd** can't run in container, **node part** of the CSI driver run as **systemd service** on the Kubernetes nodes as an alternative we call **neonsan-plugin**. **Neonsan-plugin** will ben  installed by **ansible**. 
   If you have installed kubesphere by ks-installer, ansible has been installed on the master. The path of ansible config is **/root/kubesphere-all-v2.1.0/k8s/inventory/my_cluster/host.ini**. You could copy it to **/etc/ansible/hosts**, or run ansible like **ansible -i /root/kubesphere-all-v2.1.0/k8s/inventory/my_cluster/host.ini**.

    ```bash
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
  ```bash
  ansible-playbook deploy/neonsan/plugin/neonsan-plugin-uninstall.yaml
  ```
  - Helm
    ```bash
    helm delete csi-neonsan --namespace kube-system
    ```
  - Kubectl for kubernetes 1.16
    ```bash
    kubectl delete -f deploy/neonsan/kubernetes/release/csi-neonsan-v1.2.0-k8s16.yaml
    ```
  - Kubectl for kubernetes 1.17
    ```bash
    kubectl delete -f deploy/neonsan/kubernetes/release/csi-neonsan-v1.2.0.yaml
    ```

### StorageClass Parameters
StorageClass definition [file](deploy/neonsan/example/volume/sc.yaml) shown below is used to create StorageClass object.

```
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: csi-neonsan
  provisioner: neonsan.csi.qingstor.com
  parameters:
    fsType: "ext4"
    replica: "2"
    pool: "kube"
  reclaimPolicy: Delete 
```

- `fsType`: `ext3`, `ext4`, `xfs`. Default `ext4`.
- `replica`: count of replicas (`1-3`). Default` 1`.
- `poolName`: pool of Neonsan, should not be empty. 

## Document
- [User Guide](docs/user-guide.md)

## Support
If you have any questions or suggestions, please submit an issue at [qingstor-csi](https://github.com/yunify/qingstor-csi/issues).
