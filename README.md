# DEPlayer(DExpress Player)

------

DExpress 独立研发的播放器，针对当前播放器采用简单的hls协议，造成盗链现象无法阻止的问题而产生的一款播放器。使用了DEPlayer播放器将从根本上解决视频流盗链的情况，从而保护运营者不会花费冤枉的资金费用在CDN上。


## 目录说明：

### download_encryption
	HLS协议视频下载和加密进程，这个进程主要的作用是通过配置以及调用接口等方式，从运营者提供的第三方流中拉取视频流到本地。将本地的视频流通过设置进行加密（加密方式未AES），并向外提供HLS视频流协议支持。

### dexpress player
	基于FFMPeg开发的完整的播放器SDK。此sdk从指定的地址（这里可以是通过download_encryption产生加密流或者传统的HLS非加密流）获取视频流，并进行播放。dexpress player具有码率自适应功能会根据当前网络情况，从m3u8中选择最为合适的码率视频流进行播放。





