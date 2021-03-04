package cert

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
)

var (
	DemoKeyPair  *tls.Certificate
	DemoCertPool *x509.CertPool
	DemoAddr     string
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func init() {
	cert, err := ioutil.ReadFile("cert/certificate.crt")
	check(err)
	key, err := ioutil.ReadFile("cert/privateKey.key")
	check(err)
	pair, err := tls.X509KeyPair([]byte(cert), []byte(key))
	if err != nil {
		panic(err)
	}
	DemoKeyPair = &pair
	DemoCertPool = x509.NewCertPool()
	ok := DemoCertPool.AppendCertsFromPEM([]byte(cert))
	if !ok {
		panic("bad certs")
	}
}
