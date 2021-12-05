### 加密相关 -- 非对称加密

协议使用的非对称加密方式为ECC_256, 采用ECDH实现。因ECC相关尚未形成标准, 加密解密功能“自定义实现“；此处的”自定义“在于实现方式而不在于逻辑。源于代码https://github.com/cloudflare/redoctober/blob/master/ecdh/ecdh_test.go、https://github.com/ethereum/go-ethereum/blob/master/crypto/ecies/ecies.go、中，虽有是一样的逻辑，但是拥有不一样的实现


```shell
	加密时, 自己生成一对ECC密钥, 为selfPri, selfPub; 然后使用传入的公钥计算ScalarMult(selfPri.D)可以获得随机数sk; 最后使用这个随机数作为对称加密的密钥加密数据为ct。

	解密时需要两个数据：selfPub和ct。解密时解密方通过保留的密钥与接收到的selfPub进行对称的ScalarMult计算, 能获得相同的sk。获取到sk对ct进行解密即可。
	
	因此数据ECC加密后的密文起码包含两部分：ct 和 selfPub。如果对称加密方式没有HMAC功能, 还会携带selfPub的摘要
```

以上加密解密过程中存在可变的地方：

| II                     | 说明                                           | FUDP中ECC                                          |
| ---------------------- | ---------------------------------------------- | -------------------------------------------------- |
| ECC加密的位数          | ECC支持256、384、521位的密钥                   | 规定位256位                                        |
| selfPUb序列化方式      | elliptic存在普通序列化和序列化后压缩的两种方式 | 规定压缩                                           |
| ScalarMult的取值       | ScalarMult会返回x、y两个值，需要选取一个值     | 规定选取x                                          |
| ScalarMult值的映射方法 | ScalarMult的值的长度并不总和对称加密要求的相同 | 规定sha256                                         |
| 对称加密的选取方法     | 所有的对称加密方法都可以选取                   | 规定AES_GCM_256, 对selfPub认证, nonce为12字节的0值 |
| 拼接方法               | ct和selfPub的拼接方法                          | 格式为 [ct selfPub selfPub_len(1B)]                |





Cloudflare和以太坊实现的方式各不相同，所以才自己实现。当ECC完成标准化后会采用标准化实现方案。



