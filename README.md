
# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunify/qingstor-csi)](https://goreportcard.com/report/github.com/yunify/qingstor-csi)

## Description
QingStor CSI plugin implements an interface between Container Storage Interface ([CSI](https://github.com/container-storage-interface/)) enabled Container Orchestrator (CO) and the storage of NeonSAN, which has passed [CSI sanity test](https://github.com/kubernetes-csi/csi-test). 

## Notes
- On Kubernetes v1.16, QingStor CSI v1.2.0 not supports volume snapshot management.

## Installation 
From v1.2.0, QingStor-CSI will be installed by helm. See [Helm Charts](https://github.com/kubesphere/helm-charts/tree/master/src/test/csi-neonsan) for details.

## Document
- [User Guide](docs/user-guide.md)

## Support
If you have any questions or suggestions, please submit an issue at [qingstor-csi](https://github.com/yunify/qingstor-csi/issues).
