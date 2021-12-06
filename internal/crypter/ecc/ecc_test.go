package ecc

import (
	"bytes"
	"testing"
)

func TestCrypt(t *testing.T) {

	pri, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	tp, err := MarshalPubKey(&pri.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	tk, err := MarshalPubKey(&pri.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("length privateKey: %d  length publicKey: %d", len(tp), len(tk))

	data := []byte("0123456789abcdef0123456789abcdef")

	ct, err := Encrypt(&pri.PublicKey, data)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("length  plaintext: %d  length ciphertext: %d", len(data), len(ct))

	pt, err := Decrypt(pri, ct)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, pt) {
		t.Error(data)
		t.Error(pt)
		t.Fatal("解密出错")
	}

	sign, err := Sign(pri, data)
	if err != nil {
		panic(err)
	}
	t.Logf("length  signature: %d", len(sign))

	if ok, err := Verify(&pri.PublicKey, sign, data); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("验签失败")
	}

}
