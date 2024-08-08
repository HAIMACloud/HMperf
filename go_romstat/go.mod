module romstat

go 1.18

require (
	github.com/firstrow/tcp_server v0.1.0
	github.com/go-ping/ping v1.1.0
	github.com/kbinani/screenshot v0.0.0-20210720154843-7d3a670d8329
	github.com/kirides/go-d3d v1.0.0
	github.com/shirou/gopsutil v2.21.11+incompatible
	github.com/shogo82148/androidbinary v1.0.3
	golang.org/x/image v0.0.0-20220902085622-e7cb96979f69
)

require (
	github.com/gen2brain/shm v0.0.0-20221026125803-c33c9e32b1c8 // indirect
	github.com/go-ole/go-ole v1.2.6 // indirect
	github.com/google/uuid v1.4.0 // indirect
	github.com/jezek/xgb v1.1.0 // indirect
	github.com/lxn/win v0.0.0-20210218163916-a377121e959e // indirect
	github.com/stretchr/testify v1.8.1 // indirect
	github.com/tklauser/go-sysconf v0.3.10 // indirect
	github.com/tklauser/numcpus v0.4.0 // indirect
	github.com/yusufpapurcu/wmi v1.2.2 // indirect
	golang.org/x/net v0.18.0 // indirect
	golang.org/x/sync v0.5.0 // indirect
	golang.org/x/sys v0.14.0 // indirect
)

replace github.com/kirides/go-d3d v1.0.0 => ./pkg/go-d3d/
