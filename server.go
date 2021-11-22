package fudp

//

// // verifyAct 处理握手包0的请求数据, 可以设置为权鉴
// func (f *Config) verifyAct(data []byte) uint8 {
// 	if len(data) < 1 {
// 		return 10
// 	}

// 	var url *url.URL
// 	var err error
// 	var code uint8
// 	if len(data) > 1 {
// 		url, err = url.Parse(string(data[1:]))
// 		if err != nil {
// 			return 11
// 		}
// 		if f.serverVerify != nil {
// 			code = f.serverVerify(url)
// 		}
// 		if code > 9 {
// 			return code
// 		}
// 	}

// 	switch data[0] {
// 	case 0:
// 		if f.auth&0b00000001 == 0 { // 下载
// 			return 20
// 		}
// 	case 1:
// 		if f.auth&0b00000010 == 0 { // 上传
// 			return 20
// 		}
// 	default:
// 		return 10
// 	}
// 	return 0
// }
