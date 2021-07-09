# DE-player(DExpress player)

------

DE-player 独立研发的播放器，针对当前播放器采用简单的hls协议，造成盗链现象无法阻止的问题而产生的一款播放器。使用了DE-player播放器将从根本上解决视频流盗链的情况。DE-player对视频流进行加密，加密后的视频流无法在其它播放器中进行播放，只可以在DE-player播放器中进行正常播放。保护运营者不会花费冤枉的资金费用在CDN上。

The player independently developed by DE-player is a player that uses a simple hls protocol for the current player, which causes the problem that the hotlink phenomenon cannot be prevented. Using the DE-player player will fundamentally solve the problem of video stream hotlinking. DE-player encrypts the video stream. The encrypted video stream cannot be played in other players, but can only be played normally in the DE-player player. Protect operators from spending unjustified capital costs on CDN.


## 目录说明：

### stream_encryption
	HLS协议视频下载和加密进程，stream_encryption功能有：
	HLS protocol video download and encryption process, stream_encryption functions include:
	
	1：根据用户设置拉取相应的视频流到本地。
	1：Pull the corresponding video stream to the local according to the user setting。
	
	2：根据用户设置对指定的视频流进行加密，stream_encryption采用AES加密方式对视频流进行加密。
	2：Encrypt the specified video stream according to user settings, stream_encryption uses AES 
	encryption to encrypt the video stream。

	3：提供音视频接口功能，允许用户在线新增、停止、删除视频流。
	3：Provide audio and video interface functions, allowing users to add, stop, and delete video 
	streams online。

	4：提供下载文件接口，避免用户需要安装Nginx模块来提供HLS播放功能。
	4：Provide download file interface to avoid users need to install Nginx module to provide HLS 
	playback function。

### stream_player
	基于FFMPeg开发的一款完整的播放器，此sdk可以从指定的地址中获取视频流，并进行播放。这里的音视频流既可以是通过
	stream_encryption加密过的音视频流，也可以是传统非加密的hls流。获取视频流，并进行播放。stream_player功能为：

	Based on a complete player developed by FFMpeg, this SDK can get the video stream from the 
	specified address and play it. The audio and video stream here can be either the audio and 
	video stream encrypted by stream_encryption, or the traditional non-encrypted hls stream. 
	Get the video stream and play it.
	
	1：码率自适应：对于多码率的HLS视频流来说，stream_player会根据当前带宽情况计算出最合理的视频码率，播放这个视
	频码率。
	1：Bit rate adaptive: For HLS video streams with multiple bit rates, stream_player will calculate 
	the most reasonable video bit rate according to the current bandwidth, and play this video bit rate.
	
	2：解密：对于加密的视频流（通过stream_encryption加密的视频流），stream_player可以自动调用获取秘钥接口，
	并对相应的TS音视频文件进行解密、播放。
	2：Decryption: For encrypted video streams (video streams encrypted by stream_encryption), 
	stream_player can automatically call the interface for obtaining the secret key, and decrypt 
	and play the corresponding TS audio and video files.
	
	3：自动选择播放器：对于解码性能不足的盒子（很多可以采用系统播放器进行硬解码，但无法采用Mediacodec进行解码）
	stream_player可以根据盒子的型号等进行自动选择播放器类型。
	3：Automatically select player: For boxes with insufficient decoding performance (many boxes can use 
	system players for hard decoding, but Mediacodec cannot be used for decoding), stream_player can 
	automatically select the player type according to the box model, etc.	


### 演示效果

	加密后流地址
	CCTV1：http://61.160.212.59:18085/channellist/channel1/index.m3u8

	采用dexpress_player播放的而效果
	![image](https://github.com/DExpress-dev/DEPlayer/blob/main/Image/image_cctv1.png)





