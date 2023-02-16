// Copyright (c) 2021-2023 https://www.haimacloud.com/
// SPDX-License-Identifier: MIT

package utils

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"romstat/build"
)

type RsaWriter struct {
	fpWriter    *bufio.Writer
	pubKey      *rsa.PublicKey
	currentLine string
}

func NewRsaWriter(filePath string, pubKey *string) *RsaWriter {
	rsaWriter := new(RsaWriter)
	if pubKey != nil {
		pemBlock, _ := pem.Decode([]byte(*pubKey))
		if pemBlock == nil {
			panic(errors.New("public key error"))
		}
		pubInterface, err := x509.ParsePKCS1PublicKey(pemBlock.Bytes)
		if err != nil {
			panic(err)
		}
		rsaWriter.pubKey = pubInterface
	}
	rsaWriter.initFpOutput(filePath)
	return rsaWriter
}
func (t *RsaWriter) safeWriteString(writeString string) {
	if t.fpWriter == nil {
		panic("writer object is not initialize")
	}
	_, err := t.fpWriter.WriteString(writeString)
	if err != nil {
		fmt.Println("Write error:", err.Error())
	}
}
func (t *RsaWriter) initFpOutput(outputFileName string) {
	var err error
	if CheckFileIsExist(outputFileName) {
		err := os.Remove(outputFileName)
		if err != nil {
			return
		}
	}
	fp, err := os.OpenFile(outputFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		panic(err)
	}
	t.fpWriter = bufio.NewWriter(fp)
	//file header processing
	t.safeWriteString("romstat:" + build.RomStatVersion + "\n")
	t.safeWriteString("hmp:" + build.HmFileVersion + "\n")
	if t.pubKey != nil { //If it is rsa encryption, add the encryption header version information
		t.safeWriteString("encrypt:rsa")
	} else {
		t.safeWriteString("encrypt:none")
	}
	t.safeWriteString("\n")
	_ = t.fpWriter.Flush()
}

func (t *RsaWriter) WriteString(sz string) {
	t.currentLine += sz
}

func (t *RsaWriter) Flush() {
	encryptString := t.rsaEncrypt([]byte(t.currentLine))
	t.safeWriteString(base64.StdEncoding.EncodeToString(encryptString) + "\n")
	_ = t.fpWriter.Flush()
	t.currentLine = ""
}

func (t *RsaWriter) rsaEncrypt(data []byte) []byte {
	if t.pubKey != nil {
		ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, t.pubKey, data)
		if err != nil {
			panic(err)
		}
		return ciphertext
	}
	return data
}
