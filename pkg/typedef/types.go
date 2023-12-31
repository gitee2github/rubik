// Copyright (c) Huawei Technologies Co., Ltd. 2022. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Jing Rui
// Create: 2021-04-27
// Description: This file contains default constants used in the project

// Package typedef is general used types.
package typedef

import (
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
)

// ContainerInfo represent container
type ContainerInfo struct {
	// Basic Information
	Name       string `json:"name"`
	ID         string `json:"id"`
	PodID      string `json:"podID"`
	CgroupRoot string `json:"cgroupRoot"`
	CgroupAddr string `json:"cgroupAddr"`
}

// NewContainerInfo create container info
func NewContainerInfo(container corev1.Container, podID, conID, cgroupRoot, podCgroupPath string) *ContainerInfo {
	c := ContainerInfo{
		Name:       container.Name,
		ID:         conID,
		PodID:      podID,
		CgroupRoot: cgroupRoot,
		CgroupAddr: filepath.Join(podCgroupPath, conID),
	}
	return &c
}

// CgroupPath return full cgroup path
func (ci *ContainerInfo) CgroupPath(subsys string) string {
	if ci == nil || ci.Name == "" {
		return ""
	}
	return filepath.Join(ci.CgroupRoot, subsys, ci.CgroupAddr)
}

// Clone return deepcopy object.
func (ci *ContainerInfo) Clone() *ContainerInfo {
	copy := *ci
	return &copy
}

// PodInfo represent pod
type PodInfo struct {
	// Basic Information
	Containers map[string]*ContainerInfo `json:"containers,omitempty"`
	Name       string                    `json:"name"`
	UID        string                    `json:"uid"`
	CgroupPath string                    `json:"cgroupPath"`
	Namespace  string                    `json:"namespace"`
	CgroupRoot string                    `json:"cgroupRoot"`

	// Service Information
	Offline         bool   `json:"offline"`
	CacheLimitLevel string `json:"cacheLimitLevel,omitempty"`

	// value of quota burst
	QuotaBurst int64 `json:"quotaBurst"`
}

// Clone return deepcopy object
func (pi *PodInfo) Clone() *PodInfo {
	if pi == nil {
		return nil
	}
	copy := *pi
	// deepcopy reference object
	copy.Containers = make(map[string]*ContainerInfo, len(pi.Containers))
	for _, c := range pi.Containers {
		copy.Containers[c.Name] = c.Clone()
	}
	return &copy
}

// AddContainerInfo store container info to checkpoint
func (pi *PodInfo) AddContainerInfo(containerInfo *ContainerInfo) {
	// key should not be empty
	if containerInfo.Name == "" {
		return
	}
	pi.Containers[containerInfo.Name] = containerInfo
}
