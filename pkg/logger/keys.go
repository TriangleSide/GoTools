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

// logEntryContextKeyType is the key used in the context for the log entry.
type logEntryContextKeyType string

const (
	// This is used to ensure there's no collision for the context key.
	logEntryContextKey logEntryContextKeyType = "logEntry"
)

// LogEntryKey is the key used for log entry field values.
// The declared type ensures the consistent naming of keys across the program.
type LogEntryKey string
