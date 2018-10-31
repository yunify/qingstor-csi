# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunify/qingstor-csi)](https://goreportcard.com/report/github.com/yunify/qingstor-csi)

> English | [中文](README_zh.md)
## Description
QingStor CSI plugin implements an interface between Container Storage Interface ([CSI](https://github.com/container-storage-interface/)) enabled Container Orchestrator (CO) and the storage of NeonSAN. Currently, QingCloud CSI plugin has volume manager and snapshot manager capabilities, and passed [CSI sanity test](https://github.com/kubernetes-csi/csi-test) in Kubernetes v1.12 environment.

## Installation

- [How to install NeonSAN CSI plugin in Kubernetes v1.12](docs/install_in_k8s_v1.12.md)
- [How to uninstall NeonSAN CSI plugin](docs/uninstall_neonsan.md)


### Usage

- How to create, delete and mount volume. Please reference [NeonSAN CSI plugin usage - volume](docs/usage_neonsan_volume.md).
- How to create and delete snapshot and restore volume from snapshot. Please reference [NeonSAN CSI plugin usage - snapshot](docs/usage_neonsan_snapshot.md).

## Reference

- [Depecdencies, features and options reference](docs/reference.md)

## Support
If you have any qustions or suggestions, please submit an issue at [qingstor-csi](https://github.com/yunify/qingstor-csi/issues).