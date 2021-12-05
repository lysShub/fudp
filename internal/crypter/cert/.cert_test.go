package cert_test

import (
	"bytes"
	"testing"

	"github.com/lysShub/fudp/internal/crypter/cert"
	"github.com/lysShub/fudp/internal/crypter/ecc"
)

func TestCert(t *testing.T) {

	ca, key, err := cert.CreateEccCert(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Fatal(len(key))

	pub, sign, data, err := cert.GetCertInfo(ca)
	if err != nil {
		t.Fatal(err)
	}

	if pri, err := cert.GetKeyInfo(key); err != nil {
		t.Fatal(err)
	} else {
		if sign1, err := ecc.Sign(pri, data); err != nil {
			t.Fatal(err)
		} else if !bytes.Equal(sign1, sign) {
			t.Log(sign)
			t.Log(sign1)
			t.Fatal("校验失败")
		}
	}

	if ok, err := ecc.Verify(pub, sign, data); err != nil {
		t.Fatal(err)
	} else if !ok {
		t.Fatal("校验失败")
	}

}
