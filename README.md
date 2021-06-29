# DE-player(DExpress player)

------

DExpress 独立研发的播放器，针对当前播放器采用简单的hls协议，造成盗链现象无法阻止的问题而产生的一款播放器。使用了DEPlayer播放器将从根本上解决视频流盗链的情况，从而保护运营者不会花费冤枉的资金费用在CDN上。


## 目录说明：

### stream_encryption
	HLS协议视频下载和加密进程，download_encryption功能为：
	
	1：根据用户设置拉取相应的视频流到本地。
	
	2：根据用户设置对指定的视频流进行加密，download_encryption采用AES加密方式对视频流进行加密。

	3：提供音视频接口功能，允许用户在线新增、停止、删除视频流。

	4：提供HLS协议接口，避免用户需要安装Nginx模块来提供HLS播放功能。

### dexpress_player
	基于FFMPeg开发的一款完整的播放器，此sdk可以从指定的地址中获取视频流，并进行播放。这里的音视频流既可以是通过
	download_encryption加密过的音视频流，也可以是传统非加密的hls流。
	获取视频流，并进行播放。dexpress player功能为：
	
	1：码率自适应功能：
		 对于多码率的HLS视频流来说，dexpress_player会根据当前带宽情况计算出最合理的视频码率，播放这个视频码率。
	
	2：解密功能：
		对于加密的视频流（通过download_encryption加密的视频流），dexpress_player可以自动调用获取秘钥接口，
		并对相应的TS音视频文件进行解密、播放。
	
	3：自动选择播放器功能：
		对于解码性能不足的盒子（很多盒子可以采用系统播放器进行硬解码，但是无法采用Mediacodec进行解码），
		dexpress_player可以根据盒子的型号等进行自动选择播放器类型。	





