class/4D1E55B2-F16F-11CF-88CB-001111000030

編譯
set GOOS=windows
set GOARCH=amd64
go build -mod vendor -o k8s-hid-device-plugin.exe cmd/k8s-device-plugin/main.go