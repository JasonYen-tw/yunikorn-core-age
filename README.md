## 概述

本文檔描述了 Yunikorn 新增的 `FairWithAgingNodePolicy` 策略。

## 主要更改文件
- `pkg/scheduler/objects/node.go`
- `pkg/scheduler/objects/nodesorting.go`
- `pkg/scheduler/partition.go`
- `pkg/scheduler/policies/nodesorting_policy.go`
- `pkg/scheduler/policies/sorting_policy.go`

###  測試檔案
- `pkg/scheduler/objects/node_collection_test.go`
- `pkg/scheduler/objects/nodesorting_test.go`
- `pkg/scheduler/policies/nodesorting_policy_test.go`

--- 

## 排序策略類型

目前支援三種節點排序策略：

1. `BinPackingPolicy` - 資源打包策略
2. `FairnessPolicy` - 公平性策略
3. `FairWithAgingNodePolicy` - 帶老化機制的公平性策略


### FairWithAgingNodePolicy 實現

```go
func (p fairWithAgingNodeSortingPolicy) ScoreNode(node *Node) float64 {
    // 獲取節點等待時間
    waitingTime := node.GetWaitingTime().Seconds()
    
    // 計算資源使用率
    resourceUsage := absResourceUsage(node, &p.resourceWeights)
    
    // 計算最終分數
    // 分數 = (1 - 資源使用率) + 0.3 * 等待時間
    score := (1 - resourceUsage) + 0.3 * waitingTime
    
    return score
}
```

### 資源使用率計算(使用原本Yunikorn的)

```go
func absResourceUsage(node *Node, weights *map[string]float64) float64 {
    totalWeight := 0.0
    usage := 0.0
    
    shares := node.GetResourceUsageShares()
    for k, v := range shares {
        weight := (*weights)[k]
        if weight == 0 || math.IsNaN(v) {
            continue
        }
        usage += v * weight
        totalWeight += weight
    }
    
    if totalWeight == 0 {
        return 0
    }
    return usage / totalWeight
}
```

## 修改過的函式說明

### 1. Node 結構體新增函式

```go
// 獲取節點等待時間
func (n *Node) GetWaitingTime() time.Duration {
    n.RLock()
    defer n.RUnlock()
    return n.waitingTime
}

// 設置節點等待時間
func (n *Node) SetWaitingTime(duration time.Duration) {
    n.Lock()
    defer n.Unlock()
    n.waitingTime = duration
}

// 更新節點等待時間
func (n *Node) UpdateWaitingTime() {
    n.Lock()
    defer n.Unlock()
    if n.lastAllocationTime.IsZero() {
        n.waitingTime = time.Since(n.createTime)
    } else {
        n.waitingTime = time.Since(n.lastAllocationTime)
    }
}
```

<!--
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 -->
# Apache YuniKorn - A Universal Scheduler

[![Build Status](https://github.com/apache/yunikorn-core/actions/workflows/push-master.yml/badge.svg)](https://github.com/apache/yunikorn-core/actions)
[![codecov](https://codecov.io/gh/apache/yunikorn-core/branch/master/graph/badge.svg)](https://codecov.io/gh/apache/yunikorn-core)
[![Go Report Card](https://goreportcard.com/badge/github.com/apache/yunikorn-core)](https://goreportcard.com/report/github.com/apache/yunikorn-core)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Repo Size](https://img.shields.io/github/repo-size/apache/yunikorn-core)](https://img.shields.io/github/repo-size/apache/yunikorn-core)

Apache YuniKorn is a light-weight, universal resource scheduler for container orchestrator systems.
It is created to achieve fine-grained resource sharing for various workloads efficiently on a large scale, multi-tenant,
and cloud-native environment. YuniKorn brings a unified, cross-platform, scheduling experience for mixed workloads that consist
of stateless batch workloads and stateful services. 

YuniKorn now supports K8s and can be deployed as a custom K8s scheduler. YuniKorn's architecture design also allows adding different
shim layer and adopt to different ResourceManager implementation including Apache Hadoop YARN, or any other systems.

## Get Started

See how to get started with running YuniKorn on Kubernetes, please read the documentation on [yunikorn.apache.org](http://yunikorn.apache.org/docs/).

Want to know more about the value of the YuniKorn project, and what YuniKorn can do? Here are some
[session recordings and demos](https://yunikorn.apache.org/community/events#past-conference--meetup-recordings).

## Get Involved

Please read [get involved](http://yunikorn.apache.org/community/get_involved) document if you want to discuss issues,
contribute your ideas, explore use cases, or participate the development.

If you want to contribute code to this repo, please read the [developer doc](http://yunikorn.apache.org/docs/next/developer_guide/build).
All the design docs are available [here](http://yunikorn.apache.org/docs/next/design/architecture).

## Code Structure

Apache YuniKorn project has the following git repositories:

- [yunikorn-core](https://github.com/apache/yunikorn-core/) : the scheduler brain :round_pushpin: 
- [yunikorn-k8shim](https://github.com/apache/yunikorn-k8shim) : the adaptor to Kubernetes
- [yunikorn-scheduler-interface](https://github.com/apache/yunikorn-scheduler-interface) : the common scheduling interface
- [yunikorn-web](https://github.com/apache/yunikorn-web) : the web UI
- [yunikorn-release](https://github.com/apache/yunikorn-release/): the repo manages yunikorn releases, including the helm charts
- [yunikorn-site](https://github.com/apache/yunikorn-site/): the source code for [yunikorn website](http://yunikorn.apache.org/)

The `yunikorn-core` is the brain of the scheduler, which makes placement decisions (allocate container X on node Y) according
to the builtin rich scheduling policies. Scheduler core implementation is agnostic to the underneath resource manager system.
