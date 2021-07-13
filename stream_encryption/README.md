# stream_encryption 管理端接口

> * addStream：此接口支持管理员新增下载并加密视频流。

```python
请求类型：
	POST
请求地址：
	http://****/admin/addStream
请求内容：
	{
	"channelName": "/opt/segmenter",	//频道名称
	"sourceUrl":"copy",			//频道原地址（目前只支持m3u8）
	"pushUrl":"copy",			//推送地址
	"SrcPath": "/opt/config.conf",	    	//流内容保存目录
	"key": "/data/channellist/channel",	//流加密采用的密钥
	"iv": "",				//流加密采用的向量
	"encryptionPath":"",			//加密后的文件路径
	}

```

> * deleteStream：此接口支持管理员删除视频流

```python
请求类型：
	GET
请求地址：
	http://****/admin/deleteStream?url=***

```

> * clearStream：此接口支持管理员清空视频流

```python
请求类型：
	GET
请求地址：
	http://****/admin/clearStream

```
