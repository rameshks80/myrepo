/*
Copyright 2014 The Kubernetes Authors.

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

package volume

import (
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/volume/util/fs"
)

var _ MetricsProvider = &metricsDu{}

// metricsDu represents a MetricsProvider that calculates the used and
// available Volume space by executing the "du" command and gathering
// filesystem info for the Volume path.
type metricsDu struct {
	// the directory path the volume is mounted to.
	path string
}

// NewMetricsDu creates a new metricsDu with the Volume path.
func NewMetricsDu(path string) MetricsProvider {
	return &metricsDu{path}
}

// GetMetrics calculates the volume usage and device free space by executing "du"
// and gathering filesystem info for the Volume path.
// See MetricsProvider.GetMetrics
func (md *metricsDu) GetMetrics() (*Metrics, error) {
	metrics := &Metrics{Time: metav1.Now()}
	if md.path == "" {
		return metrics, NewNoPathDefinedError()
	}

	err := md.runDu(metrics)
	if err != nil {
		return metrics, err
	}

	err = md.runFind(metrics)
	if err != nil {
		return metrics, err
	}

	err = md.getFsInfo(metrics)
	if err != nil {
		return metrics, err
	}

	return metrics, nil
}

// runDu executes the "du" command and writes the results to metrics.Used
func (md *metricsDu) runDu(metrics *Metrics) error {
	used, err := fs.Du(md.path)
	if err != nil {
		return err
	}
	metrics.Used = used
	return nil
}

// runFind executes the "find" command and writes the results to metrics.InodesUsed
func (md *metricsDu) runFind(metrics *Metrics) error {
	inodesUsed, err := fs.Find(md.path)
	if err != nil {
		return err
	}
	metrics.InodesUsed = resource.NewQuantity(inodesUsed, resource.BinarySI)
	return nil
}

// getFsInfo writes metrics.Capacity and metrics.Available from the filesystem
// info
func (md *metricsDu) getFsInfo(metrics *Metrics) error {
	available, capacity, _, inodes, inodesFree, _, err := fs.FsInfo(md.path)
	if err != nil {
		return NewFsInfoFailedError(err)
	}
	metrics.Available = resource.NewQuantity(available, resource.BinarySI)
	metrics.Capacity = resource.NewQuantity(capacity, resource.BinarySI)
	metrics.Inodes = resource.NewQuantity(inodes, resource.BinarySI)
	metrics.InodesFree = resource.NewQuantity(inodesFree, resource.BinarySI)
	return nil
}
