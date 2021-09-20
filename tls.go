package main

import (
	"io"
	"sync"
	"time"

	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"

	"golang.org/x/net/publicsuffix"
)

var (
	ca                tls.Certificate
	leaf              *x509.Certificate
	certCache         = newSyncMap()
	serialNumberLimit = new(big.Int).Lsh(big.NewInt(1), 128)
	serverPrivateKey  *ecdsa.PrivateKey
)

func listenTls() {
	var err error
	if ca, err = tls.LoadX509KeyPair(caCertPath, caKeyPath); err != nil {
		log.Fatal("Can not open ca file. ", err)
	}

	if leaf, err = x509.ParseCertificate(ca.Certificate[0]); err != nil {
		log.Fatal("Can not parse leaf cert. ", err)
	}

	if serverPrivateKey, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err != nil {
		log.Fatal("Can generate server private key. ", err)
	}

	listen, err := tls.Listen(
		"tcp4",
		":https",
		&tls.Config{GetCertificate: getCertificate},
	)
	if err != nil {
		log.Fatal("Can not start tls server. ", err)
	}

	for {
		if conn, err := listen.Accept(); err != nil {
			log.Warning("Lose a tls conn. ", err)
		} else {
			go relayTls(conn.(*tls.Conn))
		}
	}
}

func getCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	commonName, err := publicsuffix.EffectiveTLDPlusOne(info.ServerName)
	if err != nil {
		log.Warning("Can not parse common name from client hello. ", err)
		return nil, err
	}

	if cert, ok := certCache.get(commonName); ok {
		return cert.(*tls.Certificate), nil
	}

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Warning("Can generate serial number. ", err)
		return nil, err
	}

	now := time.Now()
	derBytes, err := x509.CreateCertificate(
		rand.Reader,
		&x509.Certificate{
			SerialNumber: serialNumber,
			NotBefore:    now,
			NotAfter:     now.AddDate(5, 0, 0),
			Subject: pkix.Name{
				CommonName:   commonName,
				Organization: []string{commonName},
			},
			DNSNames: []string{"*." + commonName, commonName},
		},
		leaf,
		serverPrivateKey.Public(),
		ca.PrivateKey,
	)
	if err != nil {
		log.Warning("Can not generate server cert. ", err)
		return nil, err
	}

	cert := &tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  serverPrivateKey,
	}
	certCache.add(commonName, cert)

	return cert, nil
}

func relayTls(src *tls.Conn) {
	if err := src.Handshake(); err != nil {
		log.Warning("Handshake with client failed. ", err)
		return
	}

	cache, ok := hostResolver.get(src.ConnectionState().ServerName)
	if !ok {
		log.Warningf("No ip for host: %s.", src.ConnectionState().ServerName)
		return
	}
	host := cache.(*hostInfo)

	dst, err := tls.Dial(
		"tcp4",
		host.addr,
		&tls.Config{InsecureSkipVerify: true, ServerName: host.sn},
	)
	if err != nil {
		log.Warning("Can not connect to object server. ", err)
		return
	}

	var g sync.WaitGroup
	g.Add(2)
	go transfer(dst, src, &g)
	go transfer(src, dst, &g)
	g.Wait()

	if err := dst.Close(); err != nil {
		log.Warning("Can not close object conn. ", err)
	}
	if err := src.Close(); err != nil {
		log.Warning("Can not close client conn. ", err)
	}
}

func transfer(src io.Reader, dst io.Writer, g *sync.WaitGroup) {
	if _, err := io.Copy(dst, src); err != nil {
	}
	if err := dst.(*tls.Conn).CloseWrite(); err != nil {
		log.Warning("Can not stop object writing. ", err)
	}
	if err := src.(*tls.Conn).CloseWrite(); err != nil {
		log.Warning("Can not stop client writing. ", err)
	}
	g.Done()
}
