package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	port = "8080"
)

var (
	tlscert, tlskey string
)

func main() {

	flag.StringVar(&tlscert, "tlsCertFile", "/etc/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&tlskey, "tlsKeyFile", "/etc/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")

	flag.Parse()

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		log.Println("received request")
		arResponse := v1beta1.AdmissionReview{
			Response: &v1beta1.AdmissionResponse{
				Allowed: false,
				Result: &metav1.Status{
					Message: "Keep calm and not add more crap in the cluster!",
				},
			},
		}
		json.NewEncoder(w).Encode(arResponse)
	})

	crt, _ := ioutil.ReadFile(tlscert)
	key, _ := ioutil.ReadFile(tlskey)
	log.Printf("\n\ncerts\n\n%v\n\n%v", string(crt), string(key))

	// start webhook server in new rountine
	go func() {
		if err := http.ListenAndServeTLS(fmt.Sprintf(":%v", port), tlscert, tlskey, mux); err != nil {
			log.Printf("Failed to listen and serve webhook server: %v", err)
		}
	}()

	log.Printf("Server running listening in port: %s", port)

	// listening shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	log.Print("Got shutdown signal, shutting down webhook server gracefully...")
}
