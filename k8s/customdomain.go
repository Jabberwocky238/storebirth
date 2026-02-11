package k8s

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"jabberwocky238/console/dblayer"
	"net"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type DomainStatus string

const (
	DomainStatusPending DomainStatus = "pending"
	DomainStatusSuccess DomainStatus = "success"
	DomainStatusError   DomainStatus = "error"
)

type CustomDomain struct {
	ID        int          `json:"id"`
	CDID      string       `json:"cdid"`
	Domain    string       `json:"domain"`
	Target    string       `json:"target"`
	TXTName   string       `json:"txt_name"`
	TXTValue  string       `json:"txt_value"`
	Status    DomainStatus `json:"status"`
	UserUID   string       `json:"user_uid"`
	CreatedAt time.Time    `json:"created_at"`
}

// generateVerifyToken generates a random verification token
func generateVerifyToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// NewCustomDomain creates a new custom domain verification request
func NewCustomDomain(userUID, domain, target string) (*CustomDomain, error) {
	cdid := generateVerifyToken()[:8]
	token := generateVerifyToken()
	txtName := fmt.Sprintf("_combinator-verify.%s", domain)
	txtValue := fmt.Sprintf("combinator-verify=%s", token)

	err := dblayer.CreateCustomDomain(cdid, userUID, domain, target, txtName, txtValue, string(DomainStatusPending))
	if err != nil {
		return nil, err
	}

	cd := &CustomDomain{
		CDID:      cdid,
		Domain:    domain,
		Target:    target,
		TXTName:   txtName,
		TXTValue:  txtValue,
		Status:    DomainStatusPending,
		UserUID:   userUID,
		CreatedAt: time.Now(),
	}

	return cd, nil
}

// VerifyTXT checks if the TXT record is correctly set
func (cd *CustomDomain) VerifyTXT() bool {
	records, err := net.LookupTXT(cd.TXTName)
	if err != nil {
		return false
	}
	for _, r := range records {
		if r == cd.TXTValue {
			return true
		}
	}
	return false
}

// StartVerification starts the verification loop (5s interval, 12 times max)
func (cd *CustomDomain) StartVerification() {
	go func() {
		for i := 0; i < 12; i++ {
			time.Sleep(5 * time.Second)
			if cd.VerifyTXT() {
				cd.Status = DomainStatusSuccess
				dblayer.UpdateCustomDomainStatus(cd.CDID, string(DomainStatusSuccess))
				cd.CreateIngressRoute()
				return
			}
		}
		cd.Status = DomainStatusError
		dblayer.UpdateCustomDomainStatus(cd.CDID, string(DomainStatusError))
	}()
}

// CreateIngressRoute creates an ExternalName Service and IngressRoute for the custom domain
func (cd *CustomDomain) CreateIngressRoute() error {
	if DynamicClient == nil || K8sClient == nil {
		return fmt.Errorf("k8s client not initialized")
	}

	ctx := context.Background()
	name := fmt.Sprintf("custom-domain-%s", cd.CDID)
	tlsSecretName := fmt.Sprintf("custom-domain-tls-%s", cd.CDID)

	// Create ExternalName Service pointing to target domain
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: IngressNamespace,
			Labels: map[string]string{
				"app":      "custom-domain",
				"user-uid": cd.UserUID,
			},
		},
		Spec: corev1.ServiceSpec{
			Type:         corev1.ServiceTypeExternalName,
			ExternalName: cd.Target,
		},
	}
	if _, err := K8sClient.CoreV1().Services(IngressNamespace).Create(ctx, svc, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create service failed: %w", err)
	}

	// Create cert-manager Certificate for the custom domain
	cert := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "cert-manager.io/v1",
			"kind":       "Certificate",
			"metadata": map[string]any{
				"name":      name,
				"namespace": IngressNamespace,
				"labels": map[string]any{
					"app":      "custom-domain",
					"user-uid": cd.UserUID,
				},
			},
			"spec": map[string]any{
				"secretName": tlsSecretName,
				"dnsNames":   []any{cd.Domain},
				"issuerRef": map[string]any{
					"name": "cert-issuer",
					"kind": "ClusterIssuer",
				},
			},
		},
	}
	if _, err := DynamicClient.Resource(certificateGVR).Namespace(IngressNamespace).Create(ctx, cert, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("create certificate failed: %w", err)
	}

	// Create IngressRoute
	ingressRoute := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "traefik.io/v1alpha1",
			"kind":       "IngressRoute",
			"metadata": map[string]any{
				"name":      name,
				"namespace": IngressNamespace,
				"labels": map[string]any{
					"app":      "custom-domain",
					"user-uid": cd.UserUID,
				},
			},
			"spec": map[string]any{
				"entryPoints": []any{"websecure"},
				"routes": []any{
					map[string]any{
						"match": fmt.Sprintf("Host(`%s`)", cd.Domain),
						"kind":  "Rule",
						"services": []any{
							map[string]any{
								"name": name,
								"port": 443,
							},
						},
					},
				},
				"tls": map[string]any{
					"secretName": tlsSecretName,
				},
			},
		},
	}

	_, err := DynamicClient.Resource(IngressRouteGVR).Namespace(IngressNamespace).Create(ctx, ingressRoute, metav1.CreateOptions{})
	return err
}

// GetCustomDomain returns a custom domain by CDID
func GetCustomDomain(cdid string) *CustomDomain {
	cd, err := dblayer.GetCustomDomain(cdid)
	if err != nil {
		return nil
	}
	return &CustomDomain{
		ID:        cd.ID,
		CDID:      cd.CDID,
		Domain:    cd.Domain,
		Target:    cd.Target,
		TXTName:   cd.TXTName,
		TXTValue:  cd.TXTValue,
		Status:    DomainStatus(cd.Status),
		UserUID:   cd.UserUID,
		CreatedAt: cd.CreatedAt,
	}
}

// ListCustomDomains returns all custom domains for a user
func ListCustomDomains(userUID string) []*CustomDomain {
	dbDomains, err := dblayer.ListCustomDomains(userUID)
	if err != nil {
		return nil
	}

	var result []*CustomDomain
	for _, cd := range dbDomains {
		result = append(result, &CustomDomain{
			ID:        cd.ID,
			CDID:      cd.CDID,
			Domain:    cd.Domain,
			Target:    cd.Target,
			TXTName:   cd.TXTName,
			TXTValue:  cd.TXTValue,
			Status:    DomainStatus(cd.Status),
			UserUID:   cd.UserUID,
			CreatedAt: cd.CreatedAt,
		})
	}
	return result
}

// DeleteCustomDomain deletes a custom domain, Service and IngressRoute
func DeleteCustomDomain(cdid string) error {
	// Delete from database
	if err := dblayer.DeleteCustomDomain(cdid); err != nil {
		return err
	}

	ctx := context.Background()
	name := fmt.Sprintf("custom-domain-%s", cdid)

	// Delete Service
	if K8sClient != nil {
		K8sClient.CoreV1().Services(IngressNamespace).Delete(ctx, name, metav1.DeleteOptions{})
	}

	// Delete IngressRoute
	if DynamicClient != nil {
		DynamicClient.Resource(IngressRouteGVR).Namespace(IngressNamespace).Delete(ctx, name, metav1.DeleteOptions{})
	}

	// Delete Certificate
	if DynamicClient != nil {
		DynamicClient.Resource(certificateGVR).Namespace(IngressNamespace).Delete(ctx, name, metav1.DeleteOptions{})
	}

	return nil
}
