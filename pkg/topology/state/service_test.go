package state

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	hubkubemock "github.com/traefik/hub-agent/pkg/crd/generated/client/hub/clientset/versioned/fake"
	traefikkubemock "github.com/traefik/hub-agent/pkg/crd/generated/client/traefik/clientset/versioned/fake"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kubemock "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"
)

func TestFetcher_GetServices(t *testing.T) {
	wantSvcs := map[string]*Service{
		"myService@myns": {
			Name:      "myService",
			Namespace: "myns",
			Annotations: map[string]string{
				"my.annotation": "foo",
			},
			Selector: map[string]string{
				"my.label": "foo",
			},
			Apps: []string{"jeanmich@myns"},
			Type: corev1.ServiceTypeClusterIP,
			status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							IP:       "1.2.3.4",
							Hostname: "foo.bar",
							Ports: []corev1.PortStatus{
								{
									Port:     443,
									Protocol: "TCP",
								},
							},
						},
					},
				},
			},
		},
	}
	wantNames := map[string]string{
		"myns-myService-443":   "myService@myns",
		"myns-myService-https": "myService@myns",
	}

	apps := map[string]*App{
		"jeanmich@myns": {
			Name:      "jeanmich",
			Kind:      "Deployment",
			Namespace: "myns",
			podLabels: map[string]string{
				"my.label": "foo",
			},
		},
		"mouette@myotherns": {
			Name:      "mouette",
			Kind:      "Deployment",
			Namespace: "myotherns",
			podLabels: map[string]string{
				"my.label": "foo",
			},
		},
	}

	objects := []runtime.Object{
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myService",
				Namespace: "myns",
				Annotations: map[string]string{
					"my.annotation": "foo",
				},
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Selector: map[string]string{
					"my.label": "foo",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
						Name: "https",
					},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							IP:       "1.2.3.4",
							Hostname: "foo.bar",
							Ports: []corev1.PortStatus{
								{
									Port:     443,
									Protocol: "TCP",
								},
							},
						},
					},
				},
			},
		},
	}

	kubeClient := kubemock.NewSimpleClientset(objects...)
	hubClient := hubkubemock.NewSimpleClientset()
	traefikClient := traefikkubemock.NewSimpleClientset()

	f, err := watchAll(context.Background(), kubeClient, hubClient, traefikClient, "v1.20.1", "cluster-id")
	require.NoError(t, err)

	gotSvcs, gotNames, err := f.getServices(apps)
	require.NoError(t, err)

	assert.Equal(t, wantSvcs, gotSvcs)
	assert.Equal(t, wantNames, gotNames)
}

func TestFetcher_SelectApps(t *testing.T) {
	tests := []struct {
		desc    string
		service *corev1.Service
		apps    map[string]*App
		want    []string
	}{
		{
			desc: "Select app with matching labels",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myService",
					Namespace: "myns",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Selector: map[string]string{
						"my.label": "foo",
					},
				},
			},
			apps: map[string]*App{
				"jeanmich@myns": {
					Name:      "jeanmich",
					Kind:      "Deployment",
					Namespace: "myns",
					podLabels: map[string]string{
						"my.label": "foo",
					},
				},
			},
			want: []string{"jeanmich@myns"},
		},
		{
			desc: "Select only app with matching labels",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myService",
					Namespace: "myns",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Selector: map[string]string{
						"my.label": "foo",
					},
				},
			},
			apps: map[string]*App{
				"jeanmich@myns": {
					Name:      "jeanmich",
					Kind:      "Deployment",
					Namespace: "myns",
					podLabels: map[string]string{
						"my.label": "bar",
					},
				},
			},
		},
		{
			desc: "Select only app in the same namespace",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myService",
					Namespace: "myns",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeClusterIP,
					Selector: map[string]string{
						"my.label": "foo",
					},
				},
			},
			apps: map[string]*App{
				"jeanmich@myns": {
					Name:      "jeanmich",
					Kind:      "Deployment",
					Namespace: "myns",
					podLabels: map[string]string{
						"my.label": "foo",
					},
				},
				"mouette@myns2": {
					Name:      "mouette",
					Kind:      "Deployment",
					Namespace: "myns2",
					podLabels: map[string]string{
						"my.label": "foo",
					},
				},
			},
			want: []string{"jeanmich@myns"},
		},
		{
			desc: "Ignore service selector if service is of type ExternalName",
			service: &corev1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myService",
					Namespace: "myns",
				},
				Spec: corev1.ServiceSpec{
					Type: corev1.ServiceTypeExternalName,
					Selector: map[string]string{
						"my.label": "foo",
					},
				},
			},
			apps: map[string]*App{
				"jeanmich@myns": {
					Name:      "jeanmich",
					Kind:      "Deployment",
					Namespace: "myns",
					podLabels: map[string]string{
						"my.label": "foo",
					},
				},
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.desc, func(t *testing.T) {
			t.Parallel()

			got := selectApps(test.apps, test.service)

			assert.Equal(t, test.want, got)
		})
	}
}

func TestFetcher_GetServiceLogs(t *testing.T) {
	objects := []runtime.Object{
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myService",
				Namespace: "myns",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Selector: map[string]string{
					"my.label": "foo",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
						Name: "https",
					},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							IP:       "1.2.3.4",
							Hostname: "foo.bar",
							Ports: []corev1.PortStatus{
								{
									Port:     443,
									Protocol: "TCP",
								},
							},
						},
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "myns",
				Labels: map[string]string{
					"my.label": "foo",
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "myns",
				Labels: map[string]string{
					"my.label": "foo",
				},
			},
		},
	}

	kubeClient := kubemock.NewSimpleClientset(objects...)
	kubeClient.PrependReactor("get", "pods", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		if action.GetSubresource() != "log" {
			return false, nil, nil
		}
		return true, nil, nil
	})

	hubClient := hubkubemock.NewSimpleClientset()
	traefikClient := traefikkubemock.NewSimpleClientset()

	f, err := watchAll(context.Background(), kubeClient, hubClient, traefikClient, "v1.20.1", "cluster-id")
	require.NoError(t, err)

	got, err := f.GetServiceLogs(context.Background(), "myns", "myService", 20, 200)
	require.NoError(t, err)

	assert.Equal(t, []byte("fake logs\nfake logs\n"), got)
}

func TestFetcher_GetServiceLogsHandlesTooManyPods(t *testing.T) {
	objects := []runtime.Object{
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "myService",
				Namespace: "myns",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Selector: map[string]string{
					"my.label": "foo",
				},
				Ports: []corev1.ServicePort{
					{
						Port: 443,
						Name: "https",
					},
				},
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{
							IP:       "1.2.3.4",
							Hostname: "foo.bar",
							Ports: []corev1.PortStatus{
								{
									Port:     443,
									Protocol: "TCP",
								},
							},
						},
					},
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod1",
				Namespace: "myns",
				Labels: map[string]string{
					"my.label": "foo",
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod2",
				Namespace: "myns",
				Labels: map[string]string{
					"my.label": "foo",
				},
			},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pod3",
				Namespace: "myns",
				Labels: map[string]string{
					"my.label": "foo",
				},
			},
		},
	}

	kubeClient := kubemock.NewSimpleClientset(objects...)
	kubeClient.PrependReactor("get", "pods", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		if action.GetSubresource() != "log" {
			return false, nil, nil
		}
		return true, nil, nil
	})

	hubClient := hubkubemock.NewSimpleClientset()
	traefikClient := traefikkubemock.NewSimpleClientset()

	f, err := watchAll(context.Background(), kubeClient, hubClient, traefikClient, "v1.20.1", "cluster-id")
	require.NoError(t, err)

	got, err := f.GetServiceLogs(context.Background(), "myns", "myService", 2, 200)
	require.NoError(t, err)

	assert.Equal(t, []byte("fake logs\nfake logs\n"), got)
}
