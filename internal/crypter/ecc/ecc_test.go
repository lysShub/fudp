package ecc_test

import (
	"bytes"
	"testing"

	"github.com/lysShub/fudp/internal/crypter/ecc"
)

func TestCrypt(t *testing.T) {
	pri, pub, err := ecc.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("0123456789abcdef0123456789abcdef")

	ct, err := ecc.Encrypt(pub, data)
	if err != nil {
		t.Fatal(err)
	}

	pt, err := ecc.Decrypt(pri, ct)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, pt) {
		t.Error(data)
		t.Error(pt)
		t.Fatal("解密出错")
	}

	sign, err := ecc.Sign(pri, data)
	if err != nil {
		panic(err)
	}
	if ok, err := ecc.Verify(pub, sign, data); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("验签失败")
	}

}
