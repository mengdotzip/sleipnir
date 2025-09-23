## Linux compile
Possible you need these (check first)
```
sudo apt install opencl-headers ocl-icd-opencl-dev
```
Set CGO flags for opencl
```
export CGO_CFLAGS="-I/usr/include"
export CGO_LDFLAGS="-lOpenCL"
```
Build
```
go build
```

## Compiling on linux for windows:
Because of .cl code we have to use zig in this case
```
CGO_ENABLED=1 GOOS=windows GOARCH=amd64   CC="zig cc -target x86_64-windows"   CXX="zig c++ -target x86_64-windows"   go build
```

## Windows Compile
You will have to use the zig compiler for windows
```
$env:CGO_ENABLED="1"; $env:GOOS="windows"; $env:GOARCH="amd64"; $env:CC="zig cc -target x86_64-windows"; $env:CXX="zig c++ -target x86_64-windows"; go build
```