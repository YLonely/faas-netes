// Copyright (c) Alex Ellis 2017. All rights reserved.
// Copyright 2020 OpenFaaS Author(s)
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package k8s

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// watchdogPort for the OpenFaaS function watchdog
const watchdogPort = 8080

func NewFunctionLookup(ns string, kubeClient kubernetes.Interface, config FunctionLookupConfig) *FunctionLookup {
	return &FunctionLookup{
		DefaultNamespace: ns,
		kubeClient:       kubeClient,
		config:           config,
	}
}

type FunctionLookupConfig struct {
	RetriveCount    int
	RetriveInterval time.Duration
}

type FunctionLookup struct {
	DefaultNamespace string
	kubeClient       kubernetes.Interface
	config           FunctionLookupConfig
}

func getNamespace(name, defaultNamespace string) string {
	namespace := defaultNamespace
	if strings.Contains(name, ".") {
		namespace = name[strings.LastIndexAny(name, ".")+1:]
	}
	return namespace
}

func (l *FunctionLookup) Resolve(name string) (url.URL, error) {
	var urlStr string
	functionName := name
	namespace := getNamespace(name, l.DefaultNamespace)
	if err := l.verifyNamespace(namespace); err != nil {
		return url.URL{}, err
	}

	if strings.Contains(name, ".") {
		functionName = strings.TrimSuffix(name, "."+namespace)
	}
	for i := 0; i < l.config.RetriveCount; i++ {
		svc, err := l.kubeClient.CoreV1().Endpoints(namespace).Get(context.Background(), functionName, metav1.GetOptions{})
		if err != nil {
			return url.URL{}, fmt.Errorf("error listing \"%s.%s\": %s", functionName, namespace, err.Error())
		}

		if len(svc.Subsets) == 0 || len(svc.Subsets[0].Addresses) == 0 {
			// service is not ready for request, this may happen if the service is started from checkpoint
			// we wait for it here
			log.Printf("service %s.%s is not ready, wait %d", functionName, namespace, l.config.RetriveInterval.Milliseconds())
			time.Sleep(l.config.RetriveInterval)
			continue
		}

		all := len(svc.Subsets[0].Addresses)
		target := rand.Intn(all)

		serviceIP := svc.Subsets[0].Addresses[target].IP

		urlStr = fmt.Sprintf("http://%s:%d", serviceIP, watchdogPort)

		urlRes, err := url.Parse(urlStr)
		if err != nil {
			return url.URL{}, err
		}

		return *urlRes, nil
	}
	return url.URL{}, fmt.Errorf("max status retrive count %d exceeded", l.config.RetriveCount)
}

func (l *FunctionLookup) verifyNamespace(name string) error {
	if name != "kube-system" {
		return nil
	}
	// ToDo use global namepace parse and validation
	return fmt.Errorf("namespace not allowed")
}
