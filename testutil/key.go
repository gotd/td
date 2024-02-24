package testutil

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
)

// RSAPrivateKey returns pre-generated test RSA private key.
func RSAPrivateKey() *rsa.PrivateKey {
	rawKey, err := base64.StdEncoding.DecodeString("MIIEowIBAAKCAQEAvPXkA3h/InI+o9Q1B01ysoRZGDlazlSwOMX5Q8KgLsOSKyUDYCI" +
		"07AWV60da4eAUgbI6BNE6B/vw2jH8gInEKb+0DyOKTPGv2t0mPq5a+I+C8xbXIVLwTHBm0mWaiiDQbcaQLBSzxhZ8BTa8VyMK8RO/XIGPoNSnJhf" +
		"LcKg5pmrIenzKDnlPDE2vEPWe8E84cknnmZQVRbbyae/Vqnu+XbadKXuhOro2r1Yz4n49jLHZfUVuyoSbLbYBoTBkdiadO5wCAU9edKl9Bt4LtBu" +
		"SC24MXK4WGWCuX5P7ujENLAMl8Evn5qabD5mMOlFWJUtlBp2iZQhOzJdHtsshMq25pwIDAQABAoIBAEV+zbA1FdTuXXlVZ3dbFY7wO/A7z9jIrtM" +
		"ChK1WHCF2zgBOKZKmof4YA843PQaLqh8VFF+HL6eWEju9XJdNk7ajCa7zrD6mOL3uzc0JxO1bopaS1OYtobELOdWxhoe8j8t/1rBPoNp+lHg6bER" +
		"D4BdP4vY7tD47V4ocADdbt3ArfbfQrEhpYh6kF/bju6PdsjfmkTihG8N8d4CqUSfxr930HFUdNXF72ga7XRG+pBFRAVgZQgNJJkXPx+41WBnFqmp" +
		"Sw44/fT6MeOzy1IoMibDcSZjA/PNSIWoeMxEDKV+6VnkbsiEkwAPotDFzvPm8qROra4JRfGEB+iU3FS08+9ECgYEA54BWA9IAKgeNbzKZkExkq9e" +
		"qrOt0PUA9DrfWZEr2GO+OR7yu7Mi6uhS2ZeUM/3+OTUPfQULZBg2YxPLJ/VuFe/8gPMczT/sZr3arKDgHCDuI/Ft+HQOoEvs+IdrvfC6alfUOnoW" +
		"62QjwfPzEp9UE52yDRsNWQsX4+qJTe6aWmWkCgYEA0PUSB8R1YQx1qWvNxYH9KA73tlFw+WA7GQMbunGEDhgjO4dFscA1YiFLonlqK7WLxqvCtSJ" +
		"an2g1paOQR0V6M5mpDKSeCvLAVhfE1p+z2MPXDx9l7mWRz5z4mJJIXtEqAIn2t7ZOG4MkebcTo3Qq+S92RVnzO1ZpKYS9jOyUyI8CgYA18koZCc7" +
		"P/IKQ7xGp9qNfCBrVwOiNfXK9A0oKhQ1kMi7NuMJqmzwoMLtwczfcMjVO/AoCgzlfl7uJ6an4SGOKyaERiLoEYVdS9Cxeau/4kycQ56Ez0a5Q/gs" +
		"0iHhWT+XmG/0UI8Wu3c5s0dph4doKs9bDnrFzTf7/KOSbY+6kQQKBgBImx8su8LdeerYd7EEU+qXJLxGCX5r6FgglMfpvM/Z5eE4KgS5gsQJ2O/j" +
		"ALU3gtmSqtP5BHrgsOETMQZM/YM8ssPetMSFoVvbjl7DBLMFOudbRdmxQHGt5ikrOokTCTLDBS1JIHt7a9IcyNR2E0NrWmaKKnstvxTDbHBAq2P3" +
		"XAoGBALoQXxKH/gwnri5ioL5LPiHb+SstmSEePS/FcQsuyvgV4a9r5yl+orZVQ0FVaTSKYXSx10Cugja/CqOVop2R7oKLi7HlOKeM4fL2GXID8qp" +
		"SxHZMoDAjdG9Ph1WgU7NI5Sxm70wtDos+vbpmDHvuYHmQ56ljX+5mD3T+ZjuYk7TM")
	if err != nil {
		panic(err)
	}
	k, err := x509.ParsePKCS1PrivateKey(rawKey)
	if err != nil {
		panic(err)
	}
	return k
}
