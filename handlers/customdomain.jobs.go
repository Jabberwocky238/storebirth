package handlers

import (
	"log"
	"net"

	"jabberwocky238/console/dblayer"
)

type DomainCheckJob struct{}

func (j *DomainCheckJob) Type() string { return "domain.check" }
func (j *DomainCheckJob) ID() string   { return "periodic" }

func (j *DomainCheckJob) Do() error {
	domains, err := dblayer.ListAllSuccessDomains()
	if err != nil {
		return err
	}

	for _, cd := range domains {
		records, err := net.LookupTXT(cd.TXTName)
		if err != nil {
			dblayer.UpdateCustomDomainStatus(cd.ID, "error")
			log.Printf("[domain-check] DNS lookup failed for %s: %v", cd.TXTName, err)
			continue
		}
		found := false
		for _, r := range records {
			if r == cd.TXTValue {
				found = true
				break
			}
		}
		if !found {
			dblayer.UpdateCustomDomainStatus(cd.ID, "error")
			log.Printf("[domain-check] TXT record missing for %s", cd.Domain)
		}
	}

	log.Printf("[domain-check] checked %d domains", len(domains))
	return nil
}
