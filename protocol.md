# fudp File reliable on UDP protocol



基于UDP/IP等数据包的可靠文件(夹)传输协议。设计主要用于P2P模式，亦可以基于TLS实现CS模式。

#### 一、协议包结构

​	`{payload}    {文件序号(6b-30b) 文件序号长度(2b) 偏置(1-8B) 偏置长度(3b) 类型(5b)} `

​    包头在尾部，最大13B、最小3B

- *类型*  段表示协议中不同类型的数据包，如握手、数据传输、速度控制....
- *偏置*  标定了此数据包中第一个字节的数据在原始文件中的位置(从0开始)；*偏置*  最大位宽64位，也就是协议最大支持的单文件大小为``2^65^-1`Byte
- *偏置长度*  字段表明了*偏置*  字段所占字节数；值为0表示*偏置*  占用1个字节(+1)
- *文件序号*  字段表明数据包对应的文件所在序号，使得协议直接支持文件夹的传输；序号为文件相对路径+文件名的ASCII升序排序。
- *文件序号长度*  字段表示*文件序号*  字段所占字节数。值为0表示*文件序号*  字段位宽只有6位，于*文件序号长度*  在同一字节。

<font style="color:red;font-size:larger">协议所以数据包都遵循此结构</font>



小端



#### 二、工作模式

- **P2P模式：** P2P模式中，传输密钥(对称加密密钥)直接由传输双方通过第三安全信道交换。
- **CS  模式：** CS依靠TLS实现的传输安全。

无论哪种模式，发送请求

#### 三、协议握手

握手确定一个fudp请求。握手包包头pt为0，bias为0，fi递增

###### 	握手流程



```
握手包序号0:
由Client--->Sever
包头: [0 0 0 0]
paylaod: {加密的密钥(xB)}
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
包头: [0 0 0 1]
paylaod: 
说明：
	回复表示Server接受传输密钥，Payload暂定为空
```



```
握手包序号2:
由Client--->Sever
包头: [0 0 0 2]
paylaod: [加密后的URL]
说明：
	url形如：fudp://host:port/download?metho=get&token=xxx&systen=widows；协议名、端口可以省略，端口默认时19986。参数中有部分key被保留，参考附表。
```



```
握手包序号3(加密):
由Server--->Client
包头: [0 0 0 3]
paylaod: [响应码(2B、小端)]

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

需要确保数据传输的正确，我们还需要做许多工作。

##### ⅰ、文件信息

​	用于传递文件信息，一个RTT内即可完成。包类型值为1，包偏置字段长度为0，包偏置字段为0，文件序号长度，文件序号是对应值

1. 发送文件信息，发送方-->接收方。数据包的payload为文件信息，格式为`0(1B) 文件大小(8B) 文件权限(3B) 文件路径(包含文件名、根文件夹文件名，统一使用‘/’分隔符)` 。

   

2. 回复，接收方-->发送方。payload格式为`1(1B) code(1B)`。payload为回复消息：0表示正常情况、文件不存在(将按文件权限创建)或文件大小小于4KB；1表示文件已经存在且大于4KB(进入断点续传)；...... 后续完善



​	不同系统文件命名限制不同：

​			Windows：除ASCII的控制字符和 `\/:*?"<>|`

​			Linux，OS-X：除null或 `/`

​			Mac：不能包含`:`和以`.`开头

​	协议实现时将针对系统进行处理：删除限制字符



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

