package attestations

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	attestationSuccessTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "attestation_verification_success_total",
		Help: "Total number of successfully verified attestations",
	})

	attestationFailureTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "attestation_verification_failure_total",
			Help: "Total number of failed attestation verifications by reason",
		},
		[]string{"reason"},
	)
)
