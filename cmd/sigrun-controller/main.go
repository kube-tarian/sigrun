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
	"github.com/docker/distribution/reference"
	"github.com/pkg/errors"

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
			imgInfo, err := newImageInfo(container.Image)
			if err != nil {
				log.Println(err)
				return
			}
			img := imgInfo.String()
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

func newImageInfo(image string) (*ImageInfo, error) {
	image = addDefaultDomain(image)
	ref, err := reference.Parse(image)
	if err != nil {
		return nil, errors.Wrapf(err, "bad image: %s", image)
	}

	var registry, path, name, tag, digest string
	if named, ok := ref.(reference.Named); ok {
		registry = reference.Domain(named)
		path = reference.Path(named)
		name = path[strings.LastIndex(path, "/")+1:]
	}

	if tagged, ok := ref.(reference.Tagged); ok {
		tag = tagged.Tag()
	}

	if digested, ok := ref.(reference.Digested); ok {
		digest = digested.Digest().String()
	}

	// set default tag - the domain is set via addDefaultDomain before parsing
	if tag == "" {
		tag = "latest"
	}

	return &ImageInfo{
		Registry: registry,
		Name:     name,
		Path:     path,
		Tag:      tag,
		Digest:   digest,
	}, nil
}

type ImageInfo struct {

	// Registry is the URL address of the image registry e.g. `docker.io`
	Registry string `json:"registry,omitempty"`

	// Name is the image name portion e.g. `busybox`
	Name string `json:"name"`

	// Path is the repository path and image name e.g. `some-repository/busybox`
	Path string `json:"path"`

	// Tag is the image tag e.g. `v2`
	Tag string `json:"tag,omitempty"`

	// Digest is the image digest portion e.g. `sha256:128c6e3534b842a2eec139999b8ce8aa9a2af9907e2b9269550809d18cd832a3`
	Digest string `json:"digest,omitempty"`
}

func (i *ImageInfo) String() string {
	image := i.Registry + "/" + i.Path + ":" + i.Tag
	if i.Digest != "" {
		image = image + "@" + i.Digest
	}

	return image
}

func addDefaultDomain(name string) string {
	i := strings.IndexRune(name, '/')
	if i == -1 || (!strings.ContainsAny(name[:i], ".:") && name[:i] != "localhost" && strings.ToLower(name[:i]) == name[:i]) {
		return "docker.io/" + name
	}

	return name
}
