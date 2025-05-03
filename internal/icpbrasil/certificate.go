package icpbrasil

import (
	"bytes"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"golang.org/x/crypto/ocsp"
	"golang.org/x/crypto/pkcs12"
)

// Certificate represents an ICP-Brasil certificate with its private key
type Certificate struct {
	Cert       *x509.Certificate
	PrivateKey crypto.PrivateKey
	Chain      []*x509.Certificate
}

// LoadPFX loads a PKCS#12 (.pfx/.p12) file and extracts the certificate and private key
func LoadPFX(pfxData []byte, password string) (*Certificate, error) {
	// Parse the PKCS#12 data
	privateKey, cert, err := pkcs12.Decode(pfxData, password)
	
	// Initialize an empty CA certificates slice
	var caCerts []*x509.Certificate
	if err != nil {
		return nil, fmt.Errorf("failed to decode PKCS#12 data: %w", err)
	}

	// Ensure we have a valid certificate
	if cert == nil {
		return nil, errors.New("no certificate found in PKCS#12 file")
	}

	// Ensure we have a valid private key
	if privateKey == nil {
		return nil, errors.New("no private key found in PKCS#12 file")
	}

	// Check if the private key is an RSA key
	_, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not an RSA key")
	}

	// Create the certificate chain
	chain := []*x509.Certificate{cert}
	if len(caCerts) > 0 {
		chain = append(chain, caCerts...)
	}

	return &Certificate{
		Cert:       cert,
		PrivateKey: privateKey,
		Chain:      chain,
	}, nil
}

// LoadPFXFromFile loads a PKCS#12 (.pfx/.p12) file from disk
func LoadPFXFromFile(filename string, password string) (*Certificate, error) {
	// Read the file
	pfxData, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read PKCS#12 file: %w", err)
	}

	return LoadPFX(pfxData, password)
}

// VerifyCertificate verifies the certificate against the ICP-Brasil chain
func (c *Certificate) VerifyCertificate() error {
	// Create a certificate pool with the CA certificates
	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()

	// Add all certificates except the leaf to the appropriate pool
	for i, cert := range c.Chain {
		if i > 0 { // Skip the leaf certificate
			// If it's a root certificate (self-signed)
			if bytes.Equal(cert.RawSubject, cert.RawIssuer) {
				roots.AddCert(cert)
			} else {
				intermediates.AddCert(cert)
			}
		}
	}

	// Verify the certificate
	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
		CurrentTime:   time.Now(),
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	}

	_, err := c.Cert.Verify(opts)
	if err != nil {
		return fmt.Errorf("certificate verification failed: %w", err)
	}

	return nil
}

// CheckRevocationCRL checks if the certificate is revoked using CRL
func (c *Certificate) CheckRevocationCRL() error {
	// Get CRL distribution points from the certificate
	if len(c.Cert.CRLDistributionPoints) == 0 {
		return errors.New("no CRL distribution points found in certificate")
	}

	// Try each CRL distribution point
	var lastErr error
	for _, crlDP := range c.Cert.CRLDistributionPoints {
		// Download the CRL
		resp, err := http.Get(crlDP)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		// Read the CRL data
		crlData, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		// Parse the CRL
		crl, err := x509.ParseCRL(crlData)
		if err != nil {
			lastErr = err
			continue
		}

		// Check if the certificate is in the CRL
		for _, revoked := range crl.TBSCertList.RevokedCertificates {
			if revoked.SerialNumber.Cmp(c.Cert.SerialNumber) == 0 {
				return fmt.Errorf("certificate is revoked: %s", revoked.RevocationTime)
			}
		}

		// If we got here, the certificate is not revoked according to this CRL
		return nil
	}

	// If we tried all CRLs and none worked
	if lastErr != nil {
		return fmt.Errorf("failed to check CRL: %w", lastErr)
	}

	return nil
}

// CheckRevocationOCSP checks if the certificate is revoked using OCSP
func (c *Certificate) CheckRevocationOCSP() error {
	// Get OCSP server URL from the certificate
	if len(c.Cert.OCSPServer) == 0 {
		return errors.New("no OCSP server found in certificate")
	}

	// Find the issuer certificate
	var issuer *x509.Certificate
	for _, cert := range c.Chain {
		if bytes.Equal(cert.RawSubject, c.Cert.RawIssuer) {
			issuer = cert
			break
		}
	}

	if issuer == nil {
		return errors.New("issuer certificate not found in chain")
	}

	// Create OCSP request
	ocspRequest, err := ocsp.CreateRequest(c.Cert, issuer, nil)
	if err != nil {
		return fmt.Errorf("failed to create OCSP request: %w", err)
	}

	// Send OCSP request
	ocspURL := c.Cert.OCSPServer[0]
	resp, err := http.Post(ocspURL, "application/ocsp-request", bytes.NewReader(ocspRequest))
	if err != nil {
		return fmt.Errorf("failed to send OCSP request: %w", err)
	}
	defer resp.Body.Close()

	// Read OCSP response
	ocspData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read OCSP response: %w", err)
	}

	// Parse OCSP response
	ocspResp, err := ocsp.ParseResponse(ocspData, issuer)
	if err != nil {
		return fmt.Errorf("failed to parse OCSP response: %w", err)
	}

	// Check OCSP status
	switch ocspResp.Status {
	case ocsp.Good:
		return nil
	case ocsp.Revoked:
		return fmt.Errorf("certificate is revoked: %s", ocspResp.RevokedAt)
	case ocsp.Unknown:
		return errors.New("certificate status is unknown")
	default:
		return fmt.Errorf("unexpected OCSP status: %d", ocspResp.Status)
	}
}

// ExportCertificatePEM exports the certificate as PEM
func (c *Certificate) ExportCertificatePEM() string {
	pemBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: c.Cert.Raw,
	}
	return string(pem.EncodeToMemory(pemBlock))
}

// ExportChainPEM exports the certificate chain as PEM
func (c *Certificate) ExportChainPEM() string {
	var buf bytes.Buffer
	for _, cert := range c.Chain {
		pemBlock := &pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}
		buf.Write(pem.EncodeToMemory(pemBlock))
	}
	return buf.String()
}

// IsICPBrasilCertificate checks if the certificate is from ICP-Brasil
func (c *Certificate) IsICPBrasilCertificate() bool {
	// Check for ICP-Brasil OIDs in the certificate policies
	for _, policy := range c.Cert.PolicyIdentifiers {
		// ICP-Brasil OID prefix: 2.16.76.1.x.x
		if policy.String()[:9] == "2.16.76.1" {
			return true
		}
	}
	return false
}

// GetCertificateInfo returns basic information about the certificate
func (c *Certificate) GetCertificateInfo() map[string]string {
	info := make(map[string]string)
	info["Subject"] = c.Cert.Subject.CommonName
	info["Issuer"] = c.Cert.Issuer.CommonName
	info["SerialNumber"] = c.Cert.SerialNumber.String()
	info["NotBefore"] = c.Cert.NotBefore.Format(time.RFC3339)
	info["NotAfter"] = c.Cert.NotAfter.Format(time.RFC3339)
	info["IsICPBrasil"] = fmt.Sprintf("%t", c.IsICPBrasilCertificate())
	return info
}
