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
			//response.Envs["Name"] = "name"
		}
		responses.ContainerResponses = append(responses.ContainerResponses, response)
	}

	return &responses, nil
}

func main() {

	flag.Set("alsologtostderr", "true")
	virtualHidDevice := 0
	flag.IntVar(&virtualHidDevice, "vhd", 0, "set number of virtual Hid Device")

	flag.Parse()
	devs := []*pluginapi.Device{}
	hids := hid.Enumerate(0, 0)
	pluginSocksDir := os.Getenv("PLUGIN_SOCK_DIR")
	if pluginSocksDir == "" {
		pluginSocksDir = pluginapi.DevicePluginPath
	}

	glog.Infof("%d hid devices found", len(hids))

	for i, deviceInfo := range hids {
		devs = append(devs, &pluginapi.Device{
			ID:     deviceInfo.Product + string(i),
			Health: pluginapi.Healthy,
		})
		glog.Infof("\ndevice number: %d\n  Path: %s\n  ProductID: %d\t  VendorID: %d\n  Product: %s\t  Manufacturer: %s\n  Interface: %d\n\n",
			i, deviceInfo.Path, deviceInfo.ProductID, deviceInfo.VendorID, deviceInfo.Product, deviceInfo.Manufacturer, deviceInfo.Interface)
		/*
			glog.Infof("device number: %d", i)
			glog.Infof("Path: %s", deviceInfo.Path)
			glog.Infof("ProductID: %d", deviceInfo.ProductID)
			glog.Infof("VendorID: %d", deviceInfo.VendorID)
			glog.Infof("Product: %s", deviceInfo.Product)
			glog.Infof("Manufacturer: %s", deviceInfo.Manufacturer)
			glog.Infof("Interface: %d", deviceInfo.Interface)*/
	}

	/*
		if len(os.Args) >= 2 {
			// virtualHidDeviceEnv := os.Getenv("VIRTUAL_HID_DEVICE")
			virtualHidDeviceEnv := os.Args[1]
			i, err := strconv.Atoi(virtualHidDeviceEnv)
			if err == nil {
				virtualHidDevice = i
			}
		}*/

	glog.Infof("add %d virtual Hid Device", virtualHidDevice)

	for i := 0; i < virtualHidDevice; i++ {
		virtualHidDeviceID := "virtualHidDevice" + strconv.Itoa(i)
		devs = append(devs, &pluginapi.Device{
			ID:     virtualHidDeviceID,
			Health: pluginapi.Healthy,
		})
		glog.Infof("add " + virtualHidDeviceID)
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
