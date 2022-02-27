# fudp File reliable on UDP protocol

#### 〇、概述

基于UDP的可靠文件( 夹 )传输协议，

- 小端

#### 一、协议包结构

` {payload} {偏移(8B)} {备用(4b)} {数据包类型(4b)}   `

- *偏移*  : 但payload是文件数据时( 文件数据包 )，偏移表示payload的第一字节在文件中的偏移  
- *备用*  用于今后扩展备用，当前实现时，如果数据包类型为文件数据包时，其值循环自增

小端



#### 二、工作模式

- **P2P模式：** P2P模式中，传输密钥(对称加密密钥)直接由传输双方通过第三安全信道交换。
- **CS  模式：** CS依靠TLS实现的传输安全。

无论哪种模式，发送请求

#### 三、协议握手

握手确定一个fudp请求。握手包包头pt为0，bias为0，fi递增 握手数据包类型字段值为0，备份字段值为0 偏移字段值由0自增

###### 	握手流程



```
握手包序号0:
由Client--->Sever
包头: [0 0 0]
payload: {加密的密钥(xB)}
说明：
	请求握手。
	P2P模式发送密钥加密密钥后的数据；Server将使用密钥解密，解密成功后回复握手包1
    CS模式数据部分为空, 之后建立TLS安全信道交换密钥。成功交换密钥后Server回复握手包1。
    
    TLS安全信道交换密钥的方式为：Client生成随机密钥，密钥发送至Server，Server回复密钥。

```

之后所有数据包的Payload均被加密

```
握手包序号1:
由Sever--->Client
包头: [1 0 0]
payload: 
说明：
	回复表示Server接受传输密钥，Payload暂定为空
```



```
握手包序号2:
由Client--->Sever
包头: [2 0 0]
payload: [加密后的URL]
说明：
	url形如：fudp://host:port/download?metho=get&token=xxx&systen=widows；协议名、端口可以省略，端口默认时19986。参数中有部分key被保留，参考附表。
```



```
握手包序号3(加密):
由Server--->Client
包头: [0 0 3]
payload: [响应码(2B、小端)]

说明：
	响应码对应HTTP状态码
```



<font color="#990000">握手中所有数据包大小不能超过协议MTU</font>：5120 Bytes

###### URL参数:

以下参数为保留参数

| KEY    | 说明     | 可选 | 枚举               |
| ------ | -------- | ---- | ------------------ |
| method | 请求方法 | No   | get、put           |
| system | 操作系统 | Yes  | windwos、linux、…. |
| atime  | 时间精度 | Yes  | 数字，单位毫秒     |
|        |          |      |                    |

###### 

#### 三、文件传输



##### ⅰ、文件信息

包含文件信息，包括大小(相对根路径)，文件名、权限等。。。；	

- 文件名：不同系统下，文件名要求是不同的；当出现冲突时，



用于传递文件信息，一个RTT内即可完成。包类型值为1，包偏置字段长度为0，包偏置字段为0，文件序号长度，文件序号是对应值

1. 发送文件信息，发送方-->接收方。数据包的payload为文件信息，格式为`0(1B) 文件大小(8B) 文件权限(3B) 文件路径(包含文件名、根文件夹文件名，统一使用‘/’分隔符)` 。

   

2. 回复，接收方-->发送方。payload格式为`1(1B) code(1B)`。payload为回复消息：0表示正常情况、文件不存在(将按文件权限创建)或文件大小小于4KB；1表示文件已经存在且大于4KB(进入断点续传)；...... 后续完善



​	不同系统文件命名限制不同：

​			Windows：除ASCII的控制字符和 `\/:*?"<>|`

​			Linux，OS-X：除null或 `/`

​			Mac：不能包含`:`和以`.`开头

​	协议实现时将针对系统进行处理：删除限制字符



##### ⅰ、文件同步

文件同步使得协议支持断电重传，文件同步的目的是为获得传输开始偏移；数据包类型是3，偏移字段是文件同步包序号。

**流程：**

```
Server---->Client
alias: 同步文件信息包
包头： [0 0 3]
payload: {文件大小(8B) 文件权限(4B) 文件路径}

说明：
	文件路径：统一使用“/”分割符；文件路径大小不能太长，不能超出协议包MTU。
	跨系统文件传输时，文件名要求可能冲突、处理在client完全一方。
	此数据包需要保证到达client。
```



```
Client--->Server
alias：同步hash数据包
包头: [1 0 3]
payload: [ {star(8B) end(8B) hash(8B)}.... ]

说明：
	如果client文件不存在，数据包的payload将为空。Client收到第一个同步文件信息包后将发送start为0的同步hash数据包。
	
```



```
Server--->Client
alias：同步hash通知包
包头：[2 0 3]
payload：[next(8B)]

说明：
	通知client发送next偏移开始的文件的hash。Client收到此数据包后将会发送同步hash数据包。
```



```
Server--->Client
alias: 同步结束通知包
包头：[3 0 3]
payload：[ transStart(8B) ]
	
说明：
	通知client文件同步过程结束; Server(Client)在发送(接收)到此数据包后将直接退出文件同步过程。
```

文件同步时不可靠，触发超时都将退出同步流程，而传输开始偏移将衰落为0。

可以在传输时创建临时文件解决以下两问题（暂未实现）：1. 减少文件同步时的时间、资源消耗；2. 可以避免文件读取导致超时，而进一步导致同步直接退出。













需要确保数据传输的正确，我们还需要做许多工作。





##### ⅱ、断点续传

​	断点续传功能不言而喻，需要实现的对存在数据的校验和得到开始传输的偏置两个功能。校验数据的大小固定为128kB，采用sha256；允许数据相邻摘要合并，如文件大小恰好为256KB，前后两个摘要分别为h1和h2，合并操作就是将h1和h2拼接后重新计算摘要；文件小于128KB或者剩余小于128KB部分直接忽略。包类型值为2，包偏置字段长度为0，包偏置字段为0，文件序号长度，文件序号是对应值。

​	1、发送校验消息。一个数据包可以包含多个校验信息。payload格式为`0(1B) {bias(8B) len(8B) sum(32B)}...` 一个数据包中可以有多个块。由数据相邻摘要合并操作可以得到一个hash树，优先从树的根开始发送。

​	2、回复校验信息。校验结果立即返回，payload格式为`1(1B) {bias(8B) len(8B) bool(1B)}`。对方得到回复立即动作，比如第一个块校验成功，则可以结束。

​	3、结束文件校验。当接收方得到的偏校验成功的所有结果时，发送此包；payload格式为`255(1B) start_bias(8B)`,此包说明文件中start_bias以前的数据已得到正确的传输。发送方接收到此包后开始从指定位置开始发送数据。此过程时不可靠的、存在丢包，因此引入超时机制，当接收方在75ms内没有收到新的回复信息时，根据现有的信息计算start_bias并回复此包。



##### ⅲ、数据传输

​	主要实现数据完整可靠的传输：丢包以及突发差错；丢包采用重传解决；数据包本身使用AES_GCM加密，本身具有校验功能，如果校验出错则放弃整个数据包，需要重传此数据包。包类型值为3，包偏置字段长度、包偏置字段、文件序号长度、文件序号是对应的值。

​	接收方会维护一个数组用于记录接收到的数据范围，因此可以计算出未正确收到的数据，将这些信息通过数据重发包告知发送方重新发送。

##### ⅳ、数据重发

​	通知发送方重新发送，发送方应该优先处理此信息。包类型值为4，包偏置字段长度、包偏置字段为0，文件序号长度、文件序号是对应的值。

##### ⅴ、文件结束包

​    表示一个文件发送完成。包类型值为5，包偏置字段长度、包偏置字段为0，文件序号长度、文件序号是对应的值。



#### 四、速度控制

​		速度控制基于echo， 发送一段数据后(多个数据包)，发送方会等待接收方的echo，接收到后才会继续发送，如果没有丢包则发送下一端数据，否则重新发送丢失部分数据。假如发送时间为0，此算法只能达到信道容量的一半，2路因此并行传输确保传输能达到最大速率。



带宽100MB/s 

数据段64KB

每秒echo数：



```
弃用
fudp是基于预期速度的速度控制，当发送方以固定速率发送文件时，接收速率肯定不大于发送速率；当传输速率接近信道带宽时，两者之间差距越大；通过发送速率于接收速率之间的关系来控制传输速度。

​		基于此框架的模型有两个要点：1.发送方能精确控制发送速率；2.速度控制算法的优劣

​		发送方计算出期望速度，因此接收方应该周期性地回复接收速率。包类型值为6，包偏置字段长度、包偏置字段为0、文件序号长度、文件序号是0。layload是接收速率、B/s，每100ms发送一次，同时作为心跳包。
```









#### 工作模式

1.  Server-Client工作模式

   类似于Web服务，一个Serve可以同时处理多个合法的Client请求。Server提供证书、Client校验证书，可以确保Client访问的Server不是伪造的；相对于P-P Mode，证书不是临时。

2.  Point-Point工作模式

   点对点传输模式；发送方Listen、接收方Request；发送方使用临时的自签证书，要求接收方输入此证书的公钥以完成验签；临时证书的公钥需要用户通过第三方信道进行传输；临时证书只会用于此次传输。







协议所有的数据包拥有相同的格式，但是协议握手的数据包和数据传输的数据包还是存在差别的：除协议握手中的数据包，其他数据包都会进行加密，加密方式为**AES_GCM_128**，payload加密，包头为附加消息。







```go

func (g *gcmAsm) MySeal(dst, nonce, plaintext, data []byte) []byte {
	if len(nonce) != g.nonceSize {
		panic("crypto/cipher: incorrect nonce length given to GCM")
	}
	if uint64(len(plaintext)) > ((1<<32)-2)*BlockSize {
		panic("crypto/cipher: message too large for GCM")
	}

	var counter, tagMask [gcmBlockSize]byte

	if len(nonce) == gcmStandardNonceSize {
		// Init counter to nonce||1
		copy(counter[:], nonce)
		counter[gcmBlockSize-1] = 1
	} else {
		// Otherwise counter = GHASH(nonce)
		gcmAesData(&g.productTable, nonce, &counter)
		gcmAesFinish(&g.productTable, &tagMask, &counter, uint64(len(nonce)), uint64(0))
	}

	encryptBlockAsm(len(g.ks)/4-1, &g.ks[0], &tagMask[0], &counter[0])

	var tagOut [gcmTagSize]byte
	gcmAesData(&g.productTable, data, &tagOut)

	// ret, out := sliceForAppend(dst, len(plaintext)+g.tagSize)
	// if subtleoverlap.InexactOverlap(out[:len(plaintext)], plaintext) {
	// 	panic("crypto/cipher: invalid buffer overlap")
	// }
	if len(plaintext) > 0 {
		gcmAesEnc(&g.productTable, plaintext, plaintext, &counter, &tagOut, g.ks)
	}
	gcmAesFinish(&g.productTable, &tagMask, &tagOut, uint64(len(plaintext)), uint64(len(data)))
	// copy(out[len(plaintext):], tagOut[:])

	return plaintext
}
```













```shell
```

