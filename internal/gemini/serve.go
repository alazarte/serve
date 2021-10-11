package gemini

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
)

var (
	StatusOK = "20 text/gemini\r\n"
)

func filename(r string) string {
	matchFilename := regexp.MustCompile(`gemini://alazarte.com/([a-z]+.gmi)`)
	matches := matchFilename.FindStringSubmatch(r)
	if len(matches) != 2 {
		log.Println("I don't know how to regex, so assume index.gmi")
		return "index.gmi"
	}
	return matches[1]
}

func Serve(addr string, pem string, key string, root string) error {
	certPem, err := os.ReadFile(pem)
	if err != nil {
		return err
	}
	keyPem, err := os.ReadFile(key)
	if err != nil {
		return err
	}
	cert, err := tls.X509KeyPair(certPem, keyPem)
	if err != nil {
		return err
	}
	cfg := &tls.Config{Certificates: []tls.Certificate{cert}}
	l, err := tls.Listen("tcp", addr, cfg)
	if err != nil {
		return err
	}
	for {
		buf := make([]byte, 100)
		conn, err := l.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		conn.Read(buf)
		filename := fmt.Sprintf("%s/%s", root, filename(string(buf)))
		f, err := os.Open(filename)
		if err == nil {
			if _, err := conn.Write([]byte("20 text/gemini\r\n")); err == nil {
				if _, err := io.Copy(conn, f); err == nil {
				} else {
					log.Printf("failed to write body [err=%s]", err)
				}
			} else {
				log.Printf("failed to write header [err=%s]", err)
			}
		} else {
			log.Printf("failed to open file: [err=%s]", err)
		}
		conn.Close()
		continue
	}
}
