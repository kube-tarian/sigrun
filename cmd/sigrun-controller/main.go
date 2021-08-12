package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/devopstoday11/sigrun/pkg/config"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/devopstoday11/sigrun/pkg/controller"

	corev1 "k8s.io/api/core/v1"
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

	mux := http.NewServeMux()
	mux.HandleFunc("/validate", func(w http.ResponseWriter, r *http.Request) {
		var req v1beta1.AdmissionReview
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Println(err)
			return
		}

		var pod corev1.Pod
		err = json.Unmarshal(req.Request.Object.Raw, &pod)
		if err != nil {
			log.Println(err)
			return
		}

		configMap, err := kclient.CoreV1().ConfigMaps(controller.SIGRUN_CONTROLLER_NAMESPACE).Get(context.Background(), controller.SIGRUN_CONTROLLER_CONFIG, v1.GetOptions{})
		if err != nil {
			log.Println(err)
			return
		}

		guidToRepo, imageToGuids, err := controller.ParseSigrunConfigMap(configMap)
		if err != nil {
			log.Println(err)
			return
		}

		var containers []corev1.Container
		containers = append(pod.Spec.Containers, pod.Spec.InitContainers...)
		for _, container := range containers {
			img, err := config.NormalizeImageName(container.Image)
			if err != nil {
				log.Println(err)
				return
			}

			digestStrippedImg := strings.Split(img, "@")[0]
			for _, guid := range imageToGuids[digestStrippedImg] {
				conf := config.GetVerificationConfigFromVerificationInfo(&guidToRepo[guid].VerificationInfo)
				err := conf.VerifyImage(img)
				if err != nil {
					arResponse := v1beta1.AdmissionReview{
						Response: &v1beta1.AdmissionResponse{
							Allowed: false,
							Result: &metav1.Status{
								Message: err.Error(),
							},
						},
					}
					json.NewEncoder(w).Encode(arResponse)
					w.WriteHeader(400)
					return
				}
			}
		}

		arResponse := v1beta1.AdmissionReview{
			Response: &v1beta1.AdmissionResponse{
				Allowed: true,
			},
		}
		json.NewEncoder(w).Encode(arResponse)
		w.WriteHeader(200)
	})

	log.Printf("Server running listening in port: %s", PORT)
	if err := http.ListenAndServeTLS(fmt.Sprintf(":%v", PORT), tlscert, tlskey, mux); err != nil {
		log.Printf("Failed to listen and serve webhook server: %v", err)
	}

}
