package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/devopstoday11/sigrun/pkg/controller"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const PORT = "8080"

func main() {
	var tlscert, tlskey string
	flag.StringVar(&tlscert, "tlsCertFile", "/etc/certs/tls.crt", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&tlskey, "tlsKeyFile", "/etc/certs/tls.key", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	kRestConf, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}
	kclient, err := kubernetes.NewForConfig(kRestConf)
	if err != nil {
		log.Fatal(err)
	}
	cache := controller.NewConfigMapCache(kclient)

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		arResponse, err := validateHandler(cache, r)
		if err != nil {
			switch err.(type) {
			case controller.ContainerValidationError:
				json.NewEncoder(w).Encode(v1beta1.AdmissionReview{
					Response: &v1beta1.AdmissionResponse{
						Allowed: false,
						Result: &metav1.Status{
							Message: err.Error(),
						},
					},
				})
				return
			default:
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(err.Error()))
				return
			}
		}
		json.NewEncoder(w).Encode(arResponse)
	})

	log.Printf("Server running listening in port: %s\n", PORT)
	if err := http.ListenAndServeTLS(fmt.Sprintf(":%v", PORT), tlscert, tlskey, mux); err != nil {
		log.Printf("Failed to listen and serve webhook server: %v", err)
	}
}

func validateHandler(cache *controller.ConfigMapCache, r *http.Request) (*v1beta1.AdmissionReview, error) {
	var req v1beta1.AdmissionReview
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, err
	}

	configMap, err := cache.Get()
	if err != nil {
		return nil, err
	}

	containers, err := controller.GetContainersFromResource(&req)
	if err != nil {
		return nil, err
	}

	err = controller.ValidateContainers(configMap, containers)
	if err != nil {
		return nil, err
	}

	return &v1beta1.AdmissionReview{
		Response: &v1beta1.AdmissionResponse{
			Allowed: true,
		},
	}, nil
}
