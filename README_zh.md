# QingStor-CSI

[![Build Status](https://travis-ci.org/yunify/qingstor-csi.svg?branch=master)](https://travis-ci.org/yunify/qingstor-csi)
[![Go Report Card](https://goreportcard.com/badge/github.com/yunify/qingstor-csi)](https://goreportcard.com/report/github.com/yunify/qingstor-csi)

> [English](README.md) | 中文
## 介绍

QingStor CSI 插件实现 [CSI](https://github.com/container-storage-interface/) 接口，使容器编排平台（如 Kubernetes）能够使用 NeonSAN 存储产品的资源。目前，QingStor CSI 插件实现了**存储卷管理**和**快照管理**功能，并在 Kubernetes v1.12 环境中通过了 [CSI Sanity 测试](https://github.com/kubernetes-csi/csi-test)。

## 安装

- [如何在 Kubernetes 1.12 安装 NeonSAN CSI 插件](docs/install_in_k8s_v1.12_zh.md)
- [如何卸载 NeonSAN CSI 插件](docs/uninstall_neonsan_zh.md)

## 用法
- 如何创建、删除和挂载存储卷，请参考：[NeonSAN CSI 插件用法-存储卷](docs/usage_neonsan_volume_zh.md)
- 如何创建、删除和创建基于快照的存储卷，请参考：[NeonSAN CSI 插件用法-快照](docs/usage_neonsan_snapshot_zh.md)

## 参考资料
- [依赖版本，特性，对象参数参考资料](docs/reference_zh.md)

## 支持
如果有任何问题或建议, 请在 [qingstor-csi](https://github.com/yunify/qingstor-csi/issues) 项目提 issue。
