package string_utils_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	stringutils "github.com/TriangleSide/GoBase/pkg/utils/string"
)

var _ = Describe("camel to snake", func() {
	DescribeTable("camel to snake",
		func(camel string, expectedSnake string) {
			snake := stringutils.CamelToUpperSnake(camel)
			Expect(snake).To(Equal(expectedSnake))
		},
		Entry("", "", ""),
		Entry("", "a", "A"),
		Entry("", "12345", "12345"),
		Entry("", "1a", "1A"),
		Entry("", "1aSplit", "1A_SPLIT"),
		Entry("", "1a1Split", "1A1_SPLIT"),
		Entry("", "MyCamelCase", "MY_CAMEL_CASE"),
		Entry("", "myCamelCase", "MY_CAMEL_CASE"),
		Entry("", "CAMELCase", "CAMEL_CASE"),
	)
})
