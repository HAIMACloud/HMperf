HMperf介绍
===========
* HMperf是由海马云开发的针对android云游戏的端性能测试和分析工具平台，手机无需ROOT，手机硬件和APP也无需作任何修改，即插即用。
* HMperf支持Android平台所有应用程序（游戏，APP应用，小程序，H5等），Android模拟器，云手机，云游戏等性能测试。
* Windows/Mac OS/Linux平台都支持对Android设备进行测试。
* HMperf官网地址：http://perf.haimacloud.com/
* HMperf平台使用说明：https://github.com/HAIMACloud/HMperf/blob/main/Web平台说明.md

### 目录介绍
* go_romstat 进行性能数据采集的代码目录
  * apk 通过包名获取安装包的具体信息的相关代码
  * build 编译参数
  * stat
    * data基础数据定义
    * piugins 实现采集帧率/带宽/硬件性能的相关代码
    * utils公共支持的方法定义
  * main.go程序入口
* py_start 开始执行和上传测试结果的代码目录
  * start.py 执行测试并采集测试数据
  * uperport.py 上传测试结果生成测试报告
* romstat 可执行的二进制文件


### 编译romstat
* githup地址：
* clone代码后，在根目录执行编译代码

```bash
sh build.sh


```
* 编译完成后会在当前目录生成一个HMperf二进制文件，可以直接在手机中运行

### 编译HMperf
* 添加打包执行文件说明
* 环境依赖：
1. python 3.6以上
2. 安装requests库：``` pip intsall requests```

### 安装及运行

* 操作步骤
* 添加token获取说明
1. 步骤1: 登陆海马云官网https://www.haimacloud.com/，根据您的PC平台下载对应的应用程序。
2. 步骤2：手机开启开发者模式，允许USB调试
3. 步骤3: USB连接手机，使用adb devices命令检查PC是否正常连接手机
4. 步骤4: 使用python执行py文件，开始测试
```bash
python start.py
```
5. 步骤5:   测试完成后使用python执行上传py文件
```bash
python upReport.py
```
6. 步骤6:   登陆HMperf平台查看测试结果

###  命令行参数
* 在执行start.py 和 upReport.py时，可以通过传入devices ID来指定需要测试设备

### Jank卡顿及stutter卡顿率说明
**HMperf Jank计算方法:**
* 同时满足以下两个条件，则认为是一次卡顿Jank
    1. 当前帧耗时大于前三帧平均帧耗时的2倍
    2. 当前帧耗时大于两帧电影帧耗时（1000ms/24*2=83ms）
* 同时满足以下两个条件，则认为是一次严重卡顿BJank
    1. 当前帧耗时大于前三帧平均耗时的2倍
    2. 当前帧耗时大于三帧电影帧耗时（1000ms/24+3=125ms）
* Jank/10min：平均10分钟卡顿次数
* Bjank/10min：平均10分钟严重卡顿次数
* jank算法
  1. HMperf卡顿算法与perfdog保持一致，经过我们的测试两者的卡顿率数据基本一致
  2. 注：在发生一次jank或bigjank后，在当前帧的后三帧，不会重新计算jank；即第1帧帧间隔满足条件被记为jank后，第2，3，4帧都满足jank或bigjank条件时，也不会记为jank

**帧率/卡顿率原始数据获取说明：**
* 帧率卡顿率使用的基础数据是android系统中SurfaceFlinger的上屏数据时间戳
```bash
dumpsys SurfaceFlinger --latency
```
* 在android系统中执行该命令，可以获取到当前运行程序画面渲染后的上屏时间，通过对上屏时间的计算，可以获取到两帧画面的帧间隔数据；通过对帧间隔数据的计算和处理，就可以计算出帧率和卡顿率
* 卡顿率计算方法，参见上文
* 帧率计算方法：帧间隔等于1000ms时的帧数量
