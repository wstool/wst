package app

import "crypto/x509"

type X509CertPool interface {
	AppendCertFromPEM(pem string) bool
	CertPool() *x509.CertPool
}

type x509CertPoolImpl struct {
	x509CertPool *x509.CertPool
}

func (x x509CertPoolImpl) CertPool() *x509.CertPool {
	return x.x509CertPool
}

func (x x509CertPoolImpl) AppendCertFromPEM(pem string) bool {
	return x.x509CertPool.AppendCertsFromPEM([]byte(pem))
}

func NewX509CertPool() X509CertPool {
	return &x509CertPoolImpl{
		x509CertPool: x509.NewCertPool(),
	}
}
