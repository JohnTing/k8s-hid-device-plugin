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
)

// stubAllocFunc creates and returns allocation response for the input allocate request
func allocFunc(r *pluginapi.AllocateRequest, devs map[string]pluginapi.Device) (*pluginapi.AllocateResponse, error) {
	var responses pluginapi.AllocateResponse
	for _, req := range r.ContainerRequests {
		response := &pluginapi.ContainerAllocateResponse{}
		for _, requestID := range req.DevicesIDs {

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
func getHIDHealth(deviceInfo *hid.DeviceInfo) string {
	//return pluginapi.Unhealthy
	return pluginapi.Healthy
}
*/
func main() {

	flag.Set("alsologtostderr", "true")
	flag.Parse()
	devs := []*pluginapi.Device{}
	hids := hid.Enumerate(0, 0)
	pluginSocksDir := os.Getenv("PLUGIN_SOCK_DIR")
	if pluginSocksDir == "" {
		pluginSocksDir = pluginapi.DevicePluginPath
	}

	for _, deviceInfo := range hids {
		devs = append(devs, &pluginapi.Device{
			ID:     deviceInfo.Path,
			Health: pluginapi.Healthy,
		})
		glog.Infof("deviceInfo: %s", deviceInfo)
	}

	virtualHidDeviceEnv := os.Getenv("VIRTUAL_HID_DEVICE")
	virtualHidDevice := 0
	i, err := strconv.Atoi(virtualHidDeviceEnv)
	if err == nil {
		virtualHidDevice = i
	}
	for i := 0; i < virtualHidDevice; i++ {

		devs = append(devs, &pluginapi.Device{
			ID:     "virtualHidDevice" + string(i+1),
			Health: pluginapi.Healthy,
		})
		glog.Infof("deviceInfo: virtualHidDevice" + string(i+1))
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
