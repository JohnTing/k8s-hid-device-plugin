/*
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

package main

import (
	"flag"
	"os"
	"path"
	"strconv"
	"time"

	// "github.com/aarnaud/k8s-directx-device-plugin/pkg/gpu-detection"
	"github.com/golang/glog"
	"github.com/karalabe/hid"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	dm "k8s.io/kubernetes/pkg/kubelet/cm/devicemanager"
)

const (
	// resourceName = "microsoft.com/directx"
	resourceName = "ntcu/hid"
	// deviceClass  = "class/5B45201D-F2F2-4F3B-85BB-30FF1F953599"
	deviceClass = "class/4d1e55b2-f16f-11cf-88cb-001111000030"
	// VirtualizedMultiplier : Multiply the number of devices by 'VirtualizedMultiplier' 1 = Non-virtualized, 5 = (all devices) * 5
	VirtualizedMultiplier = 5
)

// stubAllocFunc creates and returns allocation response for the input allocate request
func allocFunc(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error) {
	var responses pluginapi.AllocateResponse
	for _, req := range r.ContainerRequests {
		response := &pluginapi.ContainerAllocateResponse{}
		// for _, requestID := range req.DevicesIDs {
		for _, requestID := range req.DevicesIDs {
			/*
				gpu := gpu_detection.GetGPUInfo(requestID)
				if gpu == nil {
					return nil, fmt.Errorf("invalid allocation request with non-existing device %s", requestID)
				}

				if getGPUHealth(gpu) != pluginapi.Healthy {
					return nil, fmt.Errorf("invalid allocation request with unhealthy device: %s", requestID)
				}
			*/

			glog.Infof("requestID: %s", requestID)

			response.Devices = append(response.Devices, &pluginapi.DeviceSpec{
				HostPath:      deviceClass,
				ContainerPath: "",
				Permissions:   "",
			})
			if response.Envs == nil {
				response.Envs = make(map[string]string)
			}
			//response.Envs["DIRECTX_GPU_Name"] = gpu.Name
			//response.Envs["DIRECTX_GPU_PNPDeviceID"] = gpu.PNPDeviceID
			//response.Envs["DIRECTX_GPU_DriverVersion"] = gpu.DriverVersion
		}
		responses.ContainerResponses = append(responses.ContainerResponses, response)
	}

	return &responses, nil
}

/*
func getGPUHealth(gpu *gpu_detection.GPUInfo) string {
	if gpu.IsStatusOK() {
		return pluginapi.Healthy
	}
	return pluginapi.Unhealthy
}
*/

func getHIDHealth(deviceInfo *hid.DeviceInfo) string {
	/* var device, err = v.Open()
	var _, err = v.Open()
	if err != nil {
		return pluginapi.Unhealthy
	}
	return pluginapi.Healthy*/
	return pluginapi.Healthy
}

func main() {

	flag.Set("alsologtostderr", "true")
	flag.Parse()
	devs := []*pluginapi.Device{}
	//gpus := gpu_detection.GetGPUList()
	hids := hid.Enumerate(0, 0)
	pluginSocksDir := os.Getenv("PLUGIN_SOCK_DIR")
	if pluginSocksDir == "" {
		pluginSocksDir = pluginapi.DevicePluginPath
	}
	/*
		gpuMatchName := os.Getenv("DIRECTX_GPU_MATCH_NAME")
		if gpuMatchName == "" {
			gpuMatchName = "nvidia"
		}
	*/
	/*
		for _, gpuInfo := range gpus {
			if !gpuInfo.MatchName(gpuMatchName) {
				glog.Warningf("'%s' doesn't match  '%s', ignoring this gpu", gpuInfo.Name, gpuMatchName)
				continue
			}

			devs = append(devs, &pluginapi.Device{
				ID:     gpuInfo.PNPDeviceID,
				Health: getGPUHealth(&gpuInfo),
			})
			glog.Infof("GPU %s id: %s", gpuInfo.Name, gpuInfo.PNPDeviceID)
		}*/

	for i := 0; i < VirtualizedMultiplier; i++ {
		for _, deviceInfo := range hids {
			vnumber := ""
			if VirtualizedMultiplier != 1 {
				vnumber = "v" + strconv.Itoa(i)
			}
			devs = append(devs, &pluginapi.Device{
				// NVIDIA GeForce GTX 1660
				//ID:     gpuInfo.PNPDeviceID,
				ID: vnumber + deviceInfo.Path,
				// PCI\VEN_10DE&DEV_2184&SUBSYS_11673842&REV_A1\4&2DB3ECDA&0&0008
				// \\?\hid#vid_096e&pid_0006#6&1170e74d&0&0000#{4d1e55b2-f16f-11cf-88cb-001111000030}
				// Health: getGPUHealth(&gpuInfo),
				Health: getHIDHealth(&deviceInfo),
			})
			glog.Infof("deviceInfo: %s", deviceInfo)
		}
	}

	glog.Infof("pluginSocksDir: %s", pluginSocksDir)
	socketPath := path.Join(pluginSocksDir, "hid.sock")
	glog.Infof("socketPath: %s", socketPath)
	dp1 := dm.NewDevicePluginStub(devs, socketPath, resourceName, false)
	if err := dp1.Start(); err != nil {
		panic(err)

	}

	dp1.SetAllocFunc(allocFunc)

	// todo: when kubelet will success to autodetect socket, change the pluginSockDir to detect DEPRECATION file
	if err := dp1.Register(pluginapi.KubeletSocket, resourceName, ""); err != nil {
		panic(err)
	}

	for {
		time.Sleep(time.Second * 10)
		if _, err := os.Stat(socketPath); os.IsNotExist(err) {
			// exit if the socketPath is missing, cause by kubelet restart, we need to start again the plugin
			os.Exit(1)
		}
	}
}
