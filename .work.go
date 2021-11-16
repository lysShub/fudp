
// PP_S P-P模式、发送方
//  @path 待发送文件(夹)路径
//  @secretkey 本次点对点传输的令牌(实为证书公钥), 需要发送者通过第三方安全的信道发送给接收者
func PP_S(path string) (f *fudp, token []byte, err error) {
	f = new(fudp)

	f.mode = mode_pp
	f.auth = authSend
	if err = verifyPath(path, true); err != nil {
		return nil, nil, err
	}
	f.sendPath = formatPath(path)

	// 证书默认有效期为1d, 之后会导致握手会失败
	var key []byte
	if f.serverCert, key, err = cert.CreateEccCert(nil); err != nil {
		return nil, nil, err
	}
	if f.serverKey, err = cert.GetKeyInfo(key); err != nil {
		return nil, nil, err
	}

	var pubkey []byte
	if pubkey, _, _, err = cert.GetCertInfo(f.serverCert); err != nil {
		return nil, nil, err
	}
	return f, pubkey, nil
}

// P-P模式 接收方
// 	@secretkey 本次点对点传输的密钥(实为证书公钥), 发送方通过第三方信道发送过来
func PP_R(path string, secretkey []byte) (f *fudp, err error) {
	f = new(fudp)

	f.mode = mode_pp
	f.auth = authReceive
	if err = verifyPath(path, false); err != nil {
		return nil, err
	}
	f.receivePath = formatPath(path)

	// 校验sercretkey的合法性

	f.tocken = secretkey

	return f, nil
}

// C-S模式 Client 发送文件
// 	rootCA是自签证书时的验签证书, 为空将使用系统根证书
func CS_C_S(path string, selfRootCA ...[]byte) (f *fudp, err error) {
	f = new(fudp)

	f.mode = mode_cs
	f.auth = authSend
	if err = verifyPath(path, true); err != nil {
		return nil, err
	}
	f.sendPath = formatPath(path)

	if len(selfRootCA) != 0 {
		f.selfCert = selfRootCA[0]
	}

	return f, nil
}

// C-S模式 Client 接收文件
// 	rootCA是自签证书时的验签证书, 为空将使用系统根证书
func CS_C_R(path string, selfRootCA ...[]byte) (f *fudp, err error) {
	f = new(fudp)
	f.mode = mode_cs
	f.auth = authReceive
	if err = verifyPath(path, false); err != nil {
		return nil, err
	}
	f.receivePath = formatPath(path)

	if len(selfRootCA) != 0 {
		f.selfCert = selfRootCA[0]
	}

	return f, nil
}

// C-S模式 Server 接收文件
func CS_S_R(path string) (f *fudp, err error) {
	f = new(fudp)
	f.mode = mode_cs
	f.auth = authReceive
	if err = verifyPath(path, false); err != nil {
		return nil, err
	}
	f.receivePath = formatPath(path)
	return f, nil
}

// C-S模式 Server 发送文件
func CS_S_S(path string) (f *fudp, err error) {
	f = new(fudp)
	f.mode = mode_cs
	f.auth = authSend
	if err = verifyPath(path, true); err != nil {
		return nil, err
	}
	f.sendPath = formatPath(path)
	return f, nil
}

// C-S模式 Server 发送/接收文件
func CS_S(spath, rpath string) (f *fudp, err error) {
	f = new(fudp)
	f.mode = mode_cs
	f.auth = authAll
	if err = verifyPath(spath, true); err != nil {
		return nil, err
	}
	f.sendPath = formatPath(spath)
	if err = verifyPath(rpath, false); err != nil {
		return nil, err
	}
	f.receivePath = formatPath(rpath)
	return f, nil
}
