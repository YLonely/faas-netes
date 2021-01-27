package k8s

import (
	"fmt"
	"strings"
	"testing"

	corelister "k8s.io/client-go/listers/core/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
)

type FakeLister struct {
}

func (f FakeLister) List(selector labels.Selector) (ret []*corev1.Endpoints, err error) {
	return nil, nil
}

func (f FakeLister) Endpoints(namespace string) corelister.EndpointsNamespaceLister {

	return FakeNSLister{}
}

type FakeNSLister struct {
}

func (f FakeNSLister) List(selector labels.Selector) (ret []*corev1.Endpoints, err error) {
	return nil, nil
}

func (f FakeNSLister) Get(name string) (*corev1.Endpoints, error) {

	// make sure that we only send the function name to the lister
	if strings.Contains(name, ".") {
		return nil, fmt.Errorf("can not look up function name with a dot!")
	}

	ep := corev1.Endpoints{
		Subsets: []corev1.EndpointSubset{{
			Addresses: []corev1.EndpointAddress{{IP: "127.0.0.1"}},
		}},
	}

	return &ep, nil
}

func Test_FunctionLookup(t *testing.T) {
}
