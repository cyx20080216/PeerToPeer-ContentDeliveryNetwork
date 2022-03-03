cd DataAndRegistrationCenter/
GOPROXY=https://goproxy.io,direct go get
go build
cd ../Node/
GOPROXY=https://goproxy.io,direct go get
go build
cd ../Router/
GOPROXY=https://goproxy.io,direct go get
go build
cd ../
