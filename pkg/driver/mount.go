/*
Copyright 2019 The Kubernetes Authors.

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

package driver

import (
	"os"
	"path/filepath"

	"github.com/kubernetes-sigs/aws-ebs-csi-driver/pkg/mounter"
	mountutils "k8s.io/mount-utils"
)

// Mounter is the interface implemented by NodeMounter.
// A mix & match of functions defined in upstream libraries. (FormatAndMount
// from struct SafeFormatAndMount, PathExists from an old edition of
// mount.Interface). Define it explicitly so that it can be mocked and to
// insulate from oft-changing upstream interfaces/structs
type Mounter interface {
	mountutils.Interface

	FormatAndMount(source string, target string, fstype string, options []string) error
	IsCorruptedMnt(err error) bool
	GetDeviceNameFromMount(mountPath string) (string, int, error)
	MakeFile(path string) error
	MakeDir(path string) error
	PathExists(path string) (bool, error)
	NeedResize(devicePath string, deviceMountPath string) (bool, error)
	Unpublish(path string) error
	Unstage(path string) error
	NewResizeFs() (Resizefs, error)
}

type Resizefs interface {
	Resize(devicePath, deviceMountPath string) (bool, error)
}

// NodeMounter implements Mounter.
// A superstruct of SafeFormatAndMount.
type NodeMounter struct {
	*mountutils.SafeFormatAndMount
}

func newNodeMounter() (Mounter, error) {
	// mounter.NewSafeMounter returns a SafeFormatAndMount
	safeMounter, err := mounter.NewSafeMounter()
	if err != nil {
		return nil, err
	}
	return &NodeMounter{safeMounter}, nil
}

// DeviceIdentifier is for mocking os io functions used for the driver to
// identify an EBS volume's corresponding device (in Linux, the path under
// /dev; in Windows, the volume number) so that it can mount it. For volumes
// already mounted, see GetDeviceNameFromMount.
// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/nvme-ebs-volumes.html#identify-nvme-ebs-device
type DeviceIdentifier interface {
	Lstat(name string) (os.FileInfo, error)
	EvalSymlinks(path string) (string, error)
}

type nodeDeviceIdentifier struct{}

func newNodeDeviceIdentifier() DeviceIdentifier {
	return &nodeDeviceIdentifier{}
}

func (i *nodeDeviceIdentifier) Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

func (i *nodeDeviceIdentifier) EvalSymlinks(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}
