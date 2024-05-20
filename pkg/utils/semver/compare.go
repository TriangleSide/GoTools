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

package semver

import (
	"fmt"
	"regexp"
	"strconv"
)

const (
	semverMatchingRegex = `^[vV]?(\d+)(?:\.(\d+))?(?:\.(\d+))?([-][\w.]+)?([+][\w.]+)?$`
)

var (
	semverRegex = regexp.MustCompile(semverMatchingRegex)
)

// Compare compares two semantic version strings.
// It returns:
//
//	1 if the first version is greater.
//	-1 if the second version is greater.
//	0 if both versions are equal.
func Compare(v1, v2 string) (int, error) {
	regexMatches1 := semverRegex.FindStringSubmatch(v1)
	if regexMatches1 == nil {
		return 0, fmt.Errorf("invalid semver: %s", v1)
	}

	regexMatches2 := semverRegex.FindStringSubmatch(v2)
	if regexMatches2 == nil {
		return 0, fmt.Errorf("invalid semver: %s", v2)
	}

	for versionPartIndex := 1; versionPartIndex <= 3; versionPartIndex++ {
		if regexMatches1[versionPartIndex] != "" && regexMatches2[versionPartIndex] == "" {
			return -1, nil
		} else if regexMatches1[versionPartIndex] == "" && regexMatches2[versionPartIndex] != "" {
			return 1, nil
		} else if regexMatches1[versionPartIndex] == "" && regexMatches2[versionPartIndex] == "" {
			break
		}

		versionPartNum1, _ := strconv.Atoi(regexMatches1[versionPartIndex])
		versionPartNum2, _ := strconv.Atoi(regexMatches2[versionPartIndex])

		if versionPartNum1 > versionPartNum2 {
			return 1, nil
		} else if versionPartNum1 < versionPartNum2 {
			return -1, nil
		}
	}

	preRelease1, preRelease2 := regexMatches1[4], regexMatches2[4]
	if preRelease1 != "" && preRelease2 == "" {
		return -1, nil
	} else if preRelease1 == "" && preRelease2 != "" {
		return 1, nil
	} else if preRelease1 != preRelease2 {
		if preRelease1 > preRelease2 {
			return 1, nil
		} else {
			return -1, nil
		}
	}

	return 0, nil
}
