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

package logger

import (
	"github.com/sirupsen/logrus"
)

// CustomFormatter extends the formatter from logrus. This wrap it so we can modify some properties.
type customFormatter struct {
	logrus.JSONFormatter
}

// Format an entry to set custom properties.
func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	entry.Time = entry.Time.UTC()
	return f.JSONFormatter.Format(entry)
}
