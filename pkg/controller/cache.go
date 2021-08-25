package controller

import (
	"context"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ConfigMapCache struct {
	mu        *sync.RWMutex
	configMap *corev1.ConfigMap
	client    *kubernetes.Clientset
}

// TODO ideally should invalidate cache when config map is updated instead of every 5 seconds, well something is better than nothing
func NewConfigMapCache(client *kubernetes.Clientset) *ConfigMapCache {
	cache := &ConfigMapCache{mu: &sync.RWMutex{}, client: client}

	go func() {
		for {
			<-time.After(5 * time.Second)
			cache.mu.Lock()
			cache.configMap = nil
			cache.mu.Unlock()
		}
	}()

	return cache
}

func (c *ConfigMapCache) Get() (*corev1.ConfigMap, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.configMap != nil {
		return c.configMap.DeepCopy(), nil
	}

	configMap, err := c.client.CoreV1().ConfigMaps(SIGRUN_CONTROLLER_NAMESPACE).Get(context.Background(), SIGRUN_CONTROLLER_CONFIG, v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	c.configMap = configMap

	return configMap, nil
}
