// Copyright (c) Huawei Technologies Co., Ltd. 2021. All rights reserved.
// rubik licensed under the Mulan PSL v2.
// You can use this software according to the terms and conditions of the Mulan PSL v2.
// You may obtain a copy of Mulan PSL v2 at:
//     http://license.coscl.org.cn/MulanPSL2
// THIS SOFTWARE IS PROVIDED ON AN "AS IS" BASIS, WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO NON-INFRINGEMENT, MERCHANTABILITY OR FIT FOR A PARTICULAR
// PURPOSE.
// See the Mulan PSL v2 for more details.
// Author: Danni Xia
// Create: 2021-05-07
// Description: sync test

package sync

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// TestSync is Sync function test
func TestSync(t *testing.T) {
	if err := Sync(); err != nil {
		assert.Contains(t, err.Error(), "unable to load in-cluster configuration")
	}
}

// TestGetOfflinePodStruct is getOfflinePodStruct function test
func TestGetOfflinePodStruct(t *testing.T) {
	pod := corev1.Pod{
		TypeMeta: v1.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			UID: "podabc",
		},
		Spec: corev1.PodSpec{},
		Status: corev1.PodStatus{
			QOSClass: corev1.PodQOSGuaranteed,
		},
	}

	podQosInfo, err := getOfflinePodStruct(pod)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/podpodabc")

	pod.Status.QOSClass = corev1.PodQOSBurstable
	podQosInfo, err = getOfflinePodStruct(pod)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/burstable/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/burstable/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/burstable/podpodabc")

	pod.Status.QOSClass = corev1.PodQOSBestEffort
	podQosInfo, err = getOfflinePodStruct(pod)
	assert.NoError(t, err)
	assert.Equal(t, podQosInfo.PodID, string(pod.UID))
	assert.Equal(t, podQosInfo.CgroupPath, "kubepods/besteffort/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["cpu"], "/sys/fs/cgroup/cpu/kubepods/besteffort/podpodabc")
	assert.Equal(t, podQosInfo.FullPath["memory"], "/sys/fs/cgroup/memory/kubepods/besteffort/podpodabc")
}

// TestIsOffline is isOffline function test
func TestIsOffline(t *testing.T) {
	annotationMap := make(map[string]string, 1)
	annotationMap[priorityAnnotationKey] = "true"
	pod := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Annotations: annotationMap,
		},
	}

	is := isOffline(pod)
	assert.Equal(t, true, is)
}

// TestVerifyOfflinePods is verifyOfflinePods function test
func TestVerifyOfflinePods(t *testing.T) {
	err := os.Setenv(nodeNameEnvKey, "")
	assert.NoError(t, err)
	clientSet := &kubernetes.Clientset{}
	err = verifyOfflinePods(clientSet)
	assert.Equal(t, true, err != nil)
}