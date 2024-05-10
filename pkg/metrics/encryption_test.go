package metrics_test

import (
	"math/rand"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/crypto/symmetric"
	"intelligence/pkg/metrics"
	"intelligence/pkg/utils/ptr"
)

var _ = Describe("metrics encryption", func() {
	When("a symmetric encryptor is created", func() {
		var (
			enc *symmetric.Encryptor
		)

		BeforeEach(func() {
			var err error
			encryptionKey := "encryptionKey" + strconv.Itoa(rand.Int())
			enc, err = symmetric.New(encryptionKey)
			Expect(err).ToNot(HaveOccurred())
		})

		When("a metric payload is created", func() {
			var (
				metric *metrics.Metric
			)

			BeforeEach(func() {
				metric = &metrics.Metric{
					Namespace: "namespace",
					Scopes: map[string]string{
						"type": "value",
					},
					Measurement: ptr.Of[float32](12.34),
					Timestamp:   time.Now(),
				}
			})

			When("the metric is marshalled and encrypted", func() {
				var (
					payload []byte
				)

				BeforeEach(func() {
					var err error
					payload, err = metrics.MarshalAndEncrypt([]*metrics.Metric{metric}, enc)
					Expect(err).ToNot(HaveOccurred())
					Expect(payload).ToNot(BeNil())
				})

				It("should be able to be unmarshalled and decrypted if using the correct key", func() {
					decryptedMetrics, err := metrics.DecryptAndUnmarshal(payload, enc)
					Expect(err).ToNot(HaveOccurred())
					Expect(decryptedMetrics).To(HaveLen(1))
					Expect(decryptedMetrics[0].Namespace).To(Equal(metric.Namespace))
					Expect(decryptedMetrics[0].Scopes).To(Equal(metric.Scopes))
					Expect(*decryptedMetrics[0].Measurement).To(BeNumerically("~", *metric.Measurement, 0.001))
					Expect(decryptedMetrics[0].Timestamp.Round(time.Microsecond)).To(Equal(metric.Timestamp.Round(time.Microsecond)))
				})

				It("should fail to be unmarshalled and decrypted if using wrong key", func() {
					otherEnc, err := symmetric.New("wrong_key")
					Expect(err).ToNot(HaveOccurred())
					decryptedMetrics, err := metrics.DecryptAndUnmarshal(payload, otherEnc)
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(ContainSubstring("failed to unmarshal"))
					Expect(decryptedMetrics).To(BeNil())
				})
			})
		})

		When("nil metrics are marshalled and encrypted", func() {
			It("should unmarshal and decrypt nil bytes", func() {
				payload, err := metrics.MarshalAndEncrypt(nil, enc)
				Expect(err).ToNot(HaveOccurred())
				Expect(payload).ToNot(BeNil())
				decryptedMetrics, err := metrics.DecryptAndUnmarshal(payload, enc)
				Expect(err).ToNot(HaveOccurred())
				Expect(payload).ToNot(BeNil())
				Expect(decryptedMetrics).To(BeNil())
			})
		})

		When("an empty metrics slice is marshalled and encrypted", func() {
			It("should unmarshal and decrypt to an empty slice", func() {
				payload, err := metrics.MarshalAndEncrypt([]*metrics.Metric{}, enc)
				Expect(err).ToNot(HaveOccurred())
				Expect(payload).ToNot(BeNil())
				decryptedMetrics, err := metrics.DecryptAndUnmarshal(payload, enc)
				Expect(err).ToNot(HaveOccurred())
				Expect(payload).ToNot(BeNil())
				Expect(decryptedMetrics).ToNot(BeNil())
				Expect(decryptedMetrics).To(HaveLen(0))
			})
		})
	})
})
