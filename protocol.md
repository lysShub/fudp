# fudp File reliable on UDP protocol



**基于UDP的文件(夹可靠传输)；不可以传输流数据**；一个FUDP连接完成一次文件(夹)的传输。

#### 一、协议包结构

​	`{payload}    {文件序号(6b-30b) 文件序号长度(2b) 偏置(1-8B) 偏置长度(3b) 类型(5b)} `

​    可以发现，包头在尾部，主要是由于包头长度是动态的原因，最大13B、最小3B

- *类型*  段表示协议中不同类型的数据包，如握手、数据传输、速度控制....
- *偏置*  标定了此数据包中第一个字节的数据在原始文件中的位置(从0开始)；*偏置*  最大位宽64位，也就是协议最大支持的单文件大小为``2^65^-1`Byte
- *偏置长度*  字段表明了*偏置*  字段所占字节数；值为0表示*偏置*  占用1个字节(+1)
- *文件序号*  字段表明数据包对应的文件所在序号，使得协议直接支持文件夹的传输；序号为文件相对路径+文件名的ASCII升序排序。
- *文件序号长度*  字段表示*文件序号*  字段所占字节数。值为0表示*文件序号*  字段位宽只有6位，于*文件序号长度*  在同一字节。

<font style="color:red;font-size:larger">协议所以数据包都遵循此结构</font>



#### 二、协议握手

​		协议是CS架构的，服务端须先开启端口监听请求、默认端口为19986；客户端发起第一个数据包。S必须先启动等待握手：

​		协议握手包的包类型、包偏置字段长度、包偏置字段、文件序号长度是0，文件序号是握手包序号。

**握手如下：**

```
握手包序号0:
由Client--->Sever
包头: [0 0 0 0]
paylaod: [ACT(1B) {格式化数据}]
说明：
	请求握手。数据包payload的第一字节ACT表示Client的动作。
	ACT为0表示上传文件；
	ACT为1表示下载文件，此时格式化数据中最后一个key为path(请求资源路径)必须存在，path的值不遵守格式化约束。
	格式化数据是遵循格式约束的数据，用于传输附加信息，类似于 system=windows&mtu=1372形式字符，=和&字符是保留字。
```



```
握手包序号1:
由Sever--->Client
包头: [0 0 0 1]
paylaod: [CODE(1B) 证书]
说明：
	CODE表示Sever对此次握手的“态度”,可类比HTTP Status Code。
	Sever回复证书，支持CA或自签。
	协议规定非对称加密为ECC_256，摘要为sha256，对称加密方式为AES_GCM_256。
	由于FUDP主要由于点对点文件(夹)传输，与传统的CS架构存在区别：点对点传输中，Sever只是一个临时的角色、下一次FUDP传输中角色很可能发生了改变，此时无法常规地部署证书。最常见的场景：小红需要把a.zip发送给小明，为此次传输申请CA证书是不切实际的，或者安装自签的root证书也是不方便的。此时我们使用自签证书，并且要求通信双方通过其他信道分享公钥，此公钥将用于本数据包中证书的验签。----目的是必须验签证书而避免root证书的部署
	所以Client验签证书的优先级为：
	  如果传入公钥，则直接使用、忽略此包传输的证书
	  如果客户端自带自签root证书，则尝试使用自签root证书验签
	  最后尝试使用系统自带root证书验签
```



```
握手包序号2:
由Client--->Sever
包头: [0 0 0 2]
paylaod: [一定格式的加密后的密钥]
说明：
	负载是一定格式的加密后的密钥，既对称加密密钥--用于实践数据传输加密。一定格式是由于可能存在附带信息（对于ECC是必须要有附带信息的），需要对此数据块约定一种格式。
	Client收到正确的CODE后会发送此包，完成握手。
```

**至此完成握手**，接下来开始文件数据传输。至此所有协议包将被加密。



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

​		fudp是基于预期速度的速度控制，当发送方以固定速率发送文件时，接收速率肯定不大于发送速率；当传输速率接近信道带宽时，两者之间差距越大；通过发送速率于接收速率之间的关系来控制传输速度。

​		基于此框架的模型有两个要点：1.发送方能精确控制发送速率；2.速度控制算法的优劣

​		发送方计算出期望速度，因此接收方应该周期性地回复接收速率。包类型值为6，包偏置字段长度、包偏置字段为0、文件序号长度、文件序号是0。layload是接收速率、B/s，每100ms发送一次，同时作为心跳包。











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

