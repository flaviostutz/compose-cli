// +build kube

/*
   Copyright 2020 Docker Compose CLI authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package client

import (
	"fmt"
	"time"

	"github.com/docker/compose-cli/api/compose"
	"github.com/docker/compose-cli/utils"
	corev1 "k8s.io/api/core/v1"
)

func podToContainerSummary(pod corev1.Pod) compose.ContainerSummary {
	return compose.ContainerSummary{
		ID:      pod.GetObjectMeta().GetName(),
		Name:    pod.GetObjectMeta().GetName(),
		Service: pod.GetObjectMeta().GetLabels()[compose.ServiceTag],
		State:   string(pod.Status.Phase),
		Project: pod.GetObjectMeta().GetLabels()[compose.ProjectTag],
	}
}

func checkPodsState(services []string, pods []corev1.Pod, status string) (bool, map[string]string, error) {
	servicePods := map[string]string{}
	stateReached := true
	for _, pod := range pods {
		service := pod.Labels[compose.ServiceTag]

		if len(services) > 0 && !utils.StringContains(services, service) {
			continue
		}
		servicePods[service] = pod.Status.Message

		if status == compose.REMOVING {
			continue
		}
		if pod.Status.Phase == corev1.PodFailed {
			return false, servicePods, fmt.Errorf(pod.Status.Reason)
		}
		if status == compose.RUNNING && pod.Status.Phase != corev1.PodRunning {
			stateReached = false
		}
	}
	if status == compose.REMOVING && len(servicePods) > 0 {
		stateReached = false
	}
	return stateReached, servicePods, nil
}

// LogFunc defines a custom logger function (progress writer events)
type LogFunc func(pod string, stateReached bool, message string)

// WaitForStatusOptions hold the state pods should reach
type WaitForStatusOptions struct {
	ProjectName string
	Services    []string
	Status      string
	Timeout     *time.Duration
	Log         LogFunc
}
