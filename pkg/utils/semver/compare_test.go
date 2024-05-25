// Copyright (c) 2024 David Ouellette.
//
// All rights reserved.
//
// This software and its documentation are proprietary information of David Ouellette.
// No part of this software or its documentation may be copied, transferred, reproduced,
// distributed, modified, or disclosed without the prior written permission of David Ouellette.
//
// Unauthorized use of this software is strictly prohibited and may be subject to civil and
// criminal penalties.
//
// By using this software, you agree to abide by the terms specified herein.

package semver_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"intelligence/pkg/utils/semver"
)

var _ = Describe("semantic versioning compare", func() {
	DescribeTable("compare semantic versions",
		func(v1 string, v2 string, expected int, expectedErr string) {
			result, err := semver.Compare(v1, v2)
			if expectedErr == "" {
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expected))
			} else {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring(expectedErr))
			}
		},
		Entry("both versions are equal", "1.0.0", "1.0.0", 0, ""),
		Entry("both versions are equal", "1.0", "1.0", 0, ""),
		Entry("both versions are equal", "1", "1", 0, ""),
		Entry("first version greater (major)", "2.0.0", "1.0.0", 1, ""),
		Entry("second version greater (major)", "1.0.0", "2.0.0", -1, ""),
		Entry("first version greater (minor)", "1.2.0", "1.1.0", 1, ""),
		Entry("second version greater (minor)", "1.1.0", "1.2.0", -1, ""),
		Entry("first version greater (patch)", "1.0.1", "1.0.0", 1, ""),
		Entry("second version greater (patch)", "1.0.0", "1.0.1", -1, ""),
		Entry("version with 'v' prefix", "v1.0.0", "1.0.0", 0, ""),
		Entry("one short version (first version)", "1.0", "1.0.0", 1, ""),
		Entry("one short version (second version)", "1.0.0", "1.0", -1, ""),
		Entry("one short version (first version)", "1", "1.0.0", 1, ""),
		Entry("one short version (second version)", "1.0.0", "1", -1, ""),
		Entry("one short version with greater major (first version)", "2.0", "1.0.0", 1, ""),
		Entry("one short version with greater major  (second version)", "1.0.0", "2.0", -1, ""),
		Entry("one short version with smaller major (first version)", "1.0", "2.0.0", -1, ""),
		Entry("one short version with smaller major (second version)", "2.0.0", "1.0", 1, ""),
		Entry("one short version with greater minor (first version)", "1.1", "1.0.0", 1, ""),
		Entry("one short version with greater minor (second version)", "1.0.0", "1.1", -1, ""),
		Entry("one short version with smaller minor (first version)", "1.0", "1.1.0", -1, ""),
		Entry("one short version with smaller minor (second version)", "1.1.0", "1.0", 1, ""),
		Entry("versions with pre-release", "1.0.0-alpha", "1.0.0-beta", -1, ""),
		Entry("pre-release vs release (first lesser)", "1.0.0-alpha", "1.0.0", -1, ""),
		Entry("pre-release vs release (second lesser)", "1.0.0", "1.0.0-alpha", 1, ""),
		Entry("same versions with pre-release", "1.0.0-alpha", "1.0.0-alpha", 0, ""),
		Entry("different version lengths with pre-release", "1.0.0-alpha", "1.0-beta", -1, ""),
		Entry("different version lengths with pre-release and build meta", "1.0.0-alpha", "1.0-beta+meta", -1, ""),
		Entry("complex pre-release comparison", "1.0.0-alpha.1", "1.0.0-alpha.beta", -1, ""),
		Entry("hash in version", "1.0.0-sha.a2334f", "1.0.0", -1, ""),
		Entry("numeric pre-release comparison", "1.0.0-1", "1.0.0-2", -1, ""),
		Entry("mixed pre-release types", "1.0.0-alpha.1", "1.0.0-1", 1, ""),
		Entry("leading zeroes in versions", "1.02.0", "1.2.0", 0, ""),
		Entry("all components different", "2.3.4", "1.2.3", 1, ""),
		Entry("version with build metadata", "1.0.0+20210101", "1.0.0", 0, ""),
		Entry("different build metadata ignored", "1.0.0+20210101", "1.0.0+20211231", 0, ""),
		Entry("only build metadata differs", "1.0.0+001", "1.0.0+002", 0, ""),
		Entry("neglectable pre-release and build metadata", "1.0.0-alpha+001", "1.0.0-alpha+002", 0, ""),
		Entry("comparing longer vs shorter pre-release", "1.0.0-alpha.beta.gamma", "1.0.0-alpha.beta", 1, ""),
		Entry("case sensitivity in pre-release", "1.0.0-Alpha", "1.0.0-alpha", -1, ""),
		Entry("invalid characters in version", "1.0.0$", "1.0.0", 0, "invalid semver: 1.0.0$"),
		Entry("version 1 with space", "1.0.0 alpha", "1.0.0", 0, "invalid semver: 1.0.0 alpha"),
		Entry("version 2 with space", "1.0.0", "1.0.0 alpha", 0, "invalid semver: 1.0.0 alpha"),
		Entry("version 1 with missing part", "1..0", "1.0.0", 0, "invalid semver: 1..0"),
		Entry("version 2 with missing part", "1.0.0", "1..0", 0, "invalid semver: 1..0"),
	)
})
