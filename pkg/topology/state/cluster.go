package state

import (
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/external-dns/endpoint"
)

// Cluster describes a Cluster.
type Cluster struct {
	ID                    string                          `json:"id"`
	Namespaces            []string                        `json:"namespaces,omitempty"`
	Apps                  map[string]*App                 `json:"apps,omitempty"`
	Ingresses             map[string]*Ingress             `json:"ingresses,omitempty"`
	Services              map[string]*Service             `json:"services,omitempty"`
	IngressControllers    map[string]*IngressController   `json:"ingressControllers,omitempty"`
	ExternalDNSes         map[string]*ExternalDNS         `json:"externalDNSes,omitempty"`
	AccessControlPolicies map[string]*AccessControlPolicy `json:"accessControlPolicies,omitempty"`
}

// App is an abstraction of Deployments/ReplicaSets/DaemonSets/StatefulSets.
type App struct {
	Name          string            `json:"name"`
	Kind          string            `json:"kind"`
	Namespace     string            `json:"namespace"`
	Replicas      int               `json:"replicas"`
	ReadyReplicas int               `json:"readyReplicas"`
	Images        []string          `json:"images,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`

	podLabels map[string]string
}

// IngressController is an abstraction of Deployments/ReplicaSets/DaemonSets/StatefulSets that
// are a cluster's IngressController.
type IngressController struct {
	App

	Type           string   `json:"type"`
	IngressClasses []string `json:"ingressClasses,omitempty"`
	MetricsURLs    []string `json:"metricsURLs,omitempty"`
	PublicIPs      []string `json:"publicIPs,omitempty"`
}

// Service describes a Service.
type Service struct {
	Name      string             `json:"name"`
	Namespace string             `json:"namespace"`
	Type      corev1.ServiceType `json:"type"`
	Selector  map[string]string  `json:"selector"`
	Apps      []string           `json:"apps,omitempty"`

	status corev1.ServiceStatus
}

// Ingress describes an Ingress.
type Ingress struct {
	Name           string                `json:"name"`
	Namespace      string                `json:"namespace"`
	ClusterID      string                `json:"clusterId"`
	Controller     string                `json:"controller,omitempty"`
	Annotations    map[string]string     `json:"annotations,omitempty"`
	TLS            []netv1.IngressTLS    `json:"tls,omitempty"`
	Rules          []netv1.IngressRule   `json:"rules,omitempty"`
	DefaultService *netv1.IngressBackend `json:"defaultService,omitempty"`
	Services       []string              `json:"services,omitempty"`
}

// ExternalDNS describes an External DNS configured within a cluster.
type ExternalDNS struct {
	DNSName string           `json:"dnsName"`
	Targets endpoint.Targets `json:"targets"`
	TTL     endpoint.TTL     `json:"ttl"`
}

// AccessControlPolicy describes an Access Control Policy configured within a cluster.
type AccessControlPolicy struct {
	Name       string                         `json:"name"`
	Namespace  string                         `json:"namespace"`
	ClusterID  string                         `json:"clusterId"`
	Method     string                         `json:"method"`
	JWT        *AccessControlPolicyJWT        `json:"jwt,omitempty"`
	BasicAuth  *AccessControlPolicyBasicAuth  `json:"basicAuth,omitempty"`
	DigestAuth *AccessControlPolicyDigestAuth `json:"digestAuth,omitempty"`
}

// AccessControlPolicyJWT describes the settings for JWT authentication within an access control policy.
type AccessControlPolicyJWT struct {
	SigningSecret              string            `json:"signingSecret,omitempty"`
	SigningSecretBase64Encoded bool              `json:"signingSecretBase64Encoded"`
	PublicKey                  string            `json:"publicKey,omitempty"`
	JWKsFile                   string            `json:"jwksFile,omitempty"`
	JWKsURL                    string            `json:"jwksUrl,omitempty"`
	StripAuthorizationHeader   bool              `json:"stripAuthorizationHeader,omitempty"`
	ForwardHeaders             map[string]string `json:"forwardHeaders,omitempty"`
	TokenQueryKey              string            `json:"tokenQueryKey,omitempty"`
	Claims                     string            `json:"claims,omitempty"`
}

// AccessControlPolicyBasicAuth holds the HTTP basic authentication configuration.
type AccessControlPolicyBasicAuth struct {
	Users                    string `json:"users,omitempty"`
	Realm                    string `json:"realm,omitempty"`
	StripAuthorizationHeader bool   `json:"stripAuthorizationHeader,omitempty"`
	ForwardUsernameHeader    string `json:"forwardUsernameHeader,omitempty"`
}

// AccessControlPolicyDigestAuth holds the HTTP digest authentication configuration.
type AccessControlPolicyDigestAuth struct {
	Users                    string `json:"users,omitempty"`
	Realm                    string `json:"realm,omitempty"`
	StripAuthorizationHeader bool   `json:"stripAuthorizationHeader,omitempty"`
	ForwardUsernameHeader    string `json:"forwardUsernameHeader,omitempty"`
}
