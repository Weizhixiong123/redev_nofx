package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"strings"
)

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Error generating key:", err)
		os.Exit(1)
	}

	keyBytes := x509.MarshalPKCS1PrivateKey(key)
	pemBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyBytes,
	}
	pemBytes := pem.EncodeToMemory(pemBlock)

	// Replace windows newlines (\r\n) and unix newlines (\n)
	// First normalize to \n
	s := strings.ReplaceAll(string(pemBytes), "\r\n", "\n")
	
	// Then replace actual newlines with literal "\n"
	oneLine := strings.ReplaceAll(s, "\n", "\\n")
	fmt.Println(oneLine)
}
