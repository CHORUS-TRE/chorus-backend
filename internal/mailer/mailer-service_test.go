//go:build integration

package mailer

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	host = flag.String("host", "", "mail host")
	port = flag.Uint("port", 587, "mail port")
	user = flag.String("user", "", "user name")
	pwd  = flag.String("pwd", "", "password")
	to   = flag.String("to", "", "email recipient")
)

func TestMailer_SendWithTlsConfig(t *testing.T) {

	var from = make(map[uint64]string)
	from[1] = "from name"

	const rootPEM = `-----BEGIN CERTIFICATE-----
MIIC1zCCAb+gAwIBAgIJAITB5JPt+xDxMA0GCSqGSIb3DQEBCwUAMBcxFTATBgNV
BAMMDDM0YTI2YjRkMTI4MjAeFw0yMDA5MTQwNzExNTBaFw0zMDA5MTIwNzExNTBa
MBcxFTATBgNVBAMMDDM0YTI2YjRkMTI4MjCCASIwDQYJKoZIhvcNAQEBBQADggEP
ADCCAQoCggEBAK2/gekbu6agXArkpy0qZiSPSttlN/LRf8A7IpZ30MSEnLNjX2lQ
P8P6LvpsqIl86R1V8eeNEjvA9xF4ZVD+cLhVLzdn5Jn5EjWqb6eiU4uUSwIhQjV+
1zyaphDVFsPLGK7qsdTSwTlL+pKnP3I1Lepo72oqLp+QHVsmWOMxFF4pzEsWbj/9
Ss42Axc31tvrYHy2HaCHv4pEJ/KrX+XiSbe36pshrzizw8HrD2cVgpdij4+OXCLA
OkReG8/WYaYp9GXGm+FupcPytcUbuk3MuFHEDwmliHNw8ee/ERN0kP6Dxjojq8Yr
npdFOmM2pCfipezHTwvqLST1FNqMhBIX66UCAwEAAaMmMCQwCQYDVR0TBAIwADAX
BgNVHREEEDAOggwzNGEyNmI0ZDEyODIwDQYJKoZIhvcNAQELBQADggEBAFadqG8C
2vOAT8HG7CN8MBoAnIoiiZQopRNNMuu11/i2xwvggIEAyj/v/7261HvKVT6YAVVj
GN5hXg8qVzbcSyg1dqCJnmPoZaRZ1l7GOgTEcRLIHDDSFvZZs8hbd97QSM6Y0V5G
xnUajgNisNppnToDhSiH9EoqVc0QfdovtCKTTkutJopFWNNvJMV3NELNOzLv69cH
S5GlaymUrReAhhx5nhj0gJ4B6CpevHAO9YXxwBvbeEMcCT+cd/RXF+1Yc44Xjec0
nALYE7xf4QehhhKKPqH0/IEJUMix/mnMMxX10yDPYzDFQ9DuyoYtR8ujm/UCCwAg
oyw/fOFJdeGkYEY=
-----END CERTIFICATE-----`

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		panic("failed to parse root certificate")
	}

	tlsConfig := &tls.Config{
		Rand:                        nil,
		Time:                        nil,
		Certificates:                nil,
		NameToCertificate:           nil,
		GetCertificate:              nil,
		GetClientCertificate:        nil,
		GetConfigForClient:          nil,
		VerifyPeerCertificate:       nil,
		RootCAs:                     roots,
		NextProtos:                  nil,
		ServerName:                  "34a26b4d1282",
		ClientAuth:                  0,
		ClientCAs:                   nil,
		InsecureSkipVerify:          false,
		CipherSuites:                nil,
		PreferServerCipherSuites:    false,
		SessionTicketsDisabled:      false,
		SessionTicketKey:            [32]byte{},
		ClientSessionCache:          nil,
		MinVersion:                  0,
		MaxVersion:                  0,
		CurvePreferences:            nil,
		DynamicRecordSizingDisabled: false,
		Renegotiation:               0,
		KeyLogWriter:                nil,
	}
	var emailSubjects = make(map[uint64]map[string]string)
	m, err := NewMailerService("85f068bf57d1c6", "fab908ed90a71b", "plain", from, "localhost:25", emailSubjects, tlsConfig)
	assert.Nil(t, err)

	err = m.SendMessage(context.Background(), 1, []string{"jostoph@localhost"}, "subject", "title", "message")
	assert.Nil(t, err)
}
