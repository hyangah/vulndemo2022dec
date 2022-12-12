package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"

	"example.com/app/internal/https"
	"example.com/app/internal/remote"
)

var (
	sshAddr = flag.String("ssh_addr", ":2222", "Address the gomote SSH server should listen on")
)

func retrieveSSHKeys(ctx context.Context) (publicKey, privateKey []byte, err error) {
	return remote.SSHKeyPair()
}

func main() {

	ctx := context.Background()
	sshCA := mustRetrieveSSHCertificateAuthority()
	sp := remote.NewSessionPool(context.Background())
	mux := http.NewServeMux()

	configureSSHServer := func() (*remote.SSHServer, error) {
		privateKey, publicKey, err := retrieveSSHKeys(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve keys for SSH Server: %v", err)
		}
		return remote.NewSSHServer(*sshAddr, privateKey, publicKey, sshCA, sp)
	}
	sshServ, err := configureSSHServer()
	if err != nil {
		log.Printf("unable to configure SSH server: %s", err)
	} else {
		go func() {
			log.Printf("running SSH server on %s", *sshAddr)
			err := sshServ.ListenAndServe()
			log.Printf("SSH server ended with error: %v", err)
		}()
		defer func() {
			err := sshServ.Close()
			if err != nil {
				log.Printf("unable to close SSH server: %s", err)
			}
		}()
	}
	log.Fatalln(https.ListenAndServe(context.Background(), mux))
}

func mustRetrieveSSHCertificateAuthority() (privateKey []byte) {
	privateKey, _, err := remote.SSHKeyPair()
	if err != nil {
		log.Fatalf("unable to create SSH CA cert: %s", err)
	}
	return
}
