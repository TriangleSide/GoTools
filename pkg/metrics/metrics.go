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

package metrics

import (
	"time"
)

// A Metric is a standard of measurement used to evaluate the performance and efficiency of a system.
type Metric struct {
	Namespace   string            `json:"namespace" validate:"required"`
	Scopes      map[string]string `json:"scopes"    validate:"required"`
	Timestamp   time.Time         `json:"timestamp" validate:"required"`
	Measurement *float32          `json:"measurement,omitempty"`
}
