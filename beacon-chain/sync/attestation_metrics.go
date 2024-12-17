package sync

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	attestationVerificationSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "attestation_verification_success_total",
		Help: "The total number of successfully verified attestations",
	})
	attestationVerificationFailure = promauto.NewCounter(prometheus.CounterOpts{
		Name: "attestation_verification_failure_total",
		Help: "The total number of failed attestation verifications",
	})
	attestationVerificationFailureReasons = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "attestation_verification_failure_reasons_total",
			Help: "The total number of attestation verification failures by reason",
		},
		[]string{"reason"},
	)
	attestationVerificationLatency = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "attestation_verification_latency_seconds",
			Help:    "Latency of attestation verification in seconds",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
	)
)
