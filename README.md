# HARQ(Data Express)

------

混合自动重传请求（Hybrid Automatic Repeat reQuest，HARQ），是一种将前向纠错编码（FEC）和自动重传请求（ARQ）相结合而形成的技术。（参见：百度百科 HARQ 词条）

# DExpress(Data Express)

------

DExpress是一款低延时安全数据传输产品。这款产品采用HARQ（混合自动重传请求）技术，实现了底层数据的可靠传输，FEC冗余包控制在10%之内。同时可选择动态加密方式进行传输数据，为每次数据传输均采用了动态加密方式，保证了数据在传输过程中的安全、私密性。拥塞控制采用的拥塞算法使用了线性斜率算法和丢包算法同时进行控制，保证了在不同网络环境下发送数据占用带宽能够非常平滑的进行扩充或者收敛。


## 目录说明：

### lib
	保存着常用操作系统下的DExpress的内核库，其中包括（Linux、Windows、Android以及IOS）。

### interface
	存放着使用DExpress所需要的详细接口定义声明，其中包括不同语言的声明（C++、Go、Delphi等）。

### doc
	存放着DExpress相关的所有文档。

### demo
	存放着使用不同语言（C++、Go、Delphi）编写的Demo程序，其中目前有两个完整的Demo项目
	
	file_express：
		使用DExpress编写的文件传输Demo

	media_express:
		使用DExpress编写的视频流传输Demo

	livego_express:
		使用DExpress内核实现了livego的Forward功能（使用harq将rtmp协议转发给其它的livego服务器）

## Demo实测截图

###	说明：	

- 以下截图是我编写的一个直播Demo的截图，一个用户手机A采集画面并将采集到的画面发送给中转服务器。中转服务器接收到数据后，将画面数据转发给另外的一个用户手机B。用户手机B将接收到的数据绘制到手机中。

- 为了保证数据的连贯性，在B手机端人为设置了50毫秒的延迟。
	
- 因此如果只是做云游戏从服务端将画面传到客户端进行展现，实际的时间应该为：云游戏延迟时间 = （测量时间 - 50）/ 2;
	
第一张截图：

![image](https://github.com/DExpress-dev/DExpress/blob/main/doc/live0.jpg)

第二张截图：

![image](https://github.com/DExpress-dev/DExpress/blob/main/doc/live3.jpg)

第三张截图：

![image](https://github.com/DExpress-dev/DExpress/blob/main/doc/live8.jpg)

## 传输波形图
![image](https://github.com/DExpress-dev/DExpress/blob/main/doc/Waveform.png)

## 专利证书
![image](https://github.com/DExpress-dev/DExpress/blob/main/doc/patent.jpg)

## UDP组播公网传输架构图
![image](https://github.com/Tinachain/DExpress/blob/master/image/framework_udp.jpg)

## 性能导图
![image](https://github.com/Tinachain/DExpress/blob/master/image/performance.jpg)


