# Simple-ContentDeliveryNetwork
It is a simple CDN.

It is open sourced under the [GPLv3](https://github.com/cyx20080216/Simple-ContentDeliveryNetwork/blob/master/LICENSE) license 

It requires [fsnotify](https://github.com/fsnotify/fsnotify)
# Use build script
```bash
sh build.sh
```
# Build center
```bash
cd DataAndRegistrationCenter/
go get
go build
```
# Build node
```bash
cd Node/
go get
go build
```
# Build router
```bash
cd Router/
go get
go build
```
