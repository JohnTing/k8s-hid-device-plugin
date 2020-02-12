# hid device plugin for Kubernetes

## what is guid-devinterface-hid
https://docs.microsoft.com/en-us/windows-hardware/drivers/install/guid-devinterface-hid

## Requirements

- Windows Server 2019 1809 or above
- docker 19.03 or above
- kubelet for windows has to support device manager, PR made here :
https://github.com/kubernetes/kubernetes/pull/80917

## Build
```cmd or powershell 
set GOOS=windows
set GOARCH=amd64
go build -mod vendor -o k8s-hid-device-plugin.exe cmd/k8s-device-plugin/main.go
```
## Run


```powershell
c:\k\k8s-hid-device-plugin.exe
```

Available environments variables :
- `PLUGIN_SOCK_DIR`  default value is `c:\var\lib\kubelet\device-plugins\`
- `vhd` is number of virtual Hid Device settings, default value is `0`
- The total number is the detected hid devices + vhd


## How to use
You can now request resources of type ntcu/hid in the container definition, the plugin will automatically add class/4D1E55B2-F16F-11CF-88CB-001111000030 as a container device

```yaml
...
spec:
  containers:
...
    resources:
      requests:
        ntcu/hid: "1"
...
```

## Links

- https://docs.microsoft.com/en-us/virtualization/windowscontainers/deploy-containers/gpu-acceleration
- https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/device-plugins/
- https://techcommunity.microsoft.com/t5/Containers/Bringing-GPU-acceleration-to-Windows-containers/ba-p/393939
- https://github.com/aarnaud/k8s-directx-device-plugin

