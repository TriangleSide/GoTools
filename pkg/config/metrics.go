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

package config

import (
	"intelligence/pkg/config/envprocessor"
)

const (
	MetricsPortEnvName           envprocessor.EnvName = "METRICS_PORT"
	MetricsKeyEnvName            envprocessor.EnvName = "METRICS_KEY"
	MetricsHostEnvName           envprocessor.EnvName = "METRICS_HOST"
	MetricsBindIPEnvName         envprocessor.EnvName = "METRICS_BIND_IP"
	MetricsQueueSizeEnvName      envprocessor.EnvName = "METRICS_QUEUE_SIZE"
	MetricsOsBufferSizeEnvName   envprocessor.EnvName = "METRICS_OS_BUFFER_SIZE"
	MetricsReadBufferSizeEnvName envprocessor.EnvName = "METRICS_READ_BUFFER_SIZE"
	MetricsReadThreadsEnvName    envprocessor.EnvName = "METRICS_READ_THREADS"
)

// MetricsCommon contains the common configuration for both the client and the server.
type MetricsCommon struct {
	MetricsPort uint16 `config_format:"snake" config_default:"35715" validate:"required,gt=0"`
	MetricsKey  string `config_format:"snake" validate:"required"`
}

// MetricsClient contains the values needed to configure a metric client.
type MetricsClient struct {
	MetricsCommon
	MetricsHost string `config_format:"snake" config_default:"::1" validate:"required"`
}

// MetricsServer contains the values needed to configure a metric server.
type MetricsServer struct {
	MetricsCommon
	MetricsBindIP         string `config_format:"snake" config_default:"::1"     validate:"required,ip_addr"`
	MetricsQueueSize      uint   `config_format:"snake" config_default:"1024"    validate:"required,gt=0,lte=8192"`
	MetricsOSBufferSize   uint   `config_format:"snake" config_default:"1048576" validate:"required,gte=4096,lte=1073741824"`
	MetricsReadBufferSize uint   `config_format:"snake" config_default:"1048576" validate:"required,gte=4096,lte=1073741824"`
	MetricsReadThreads    int    `config_format:"snake" config_default:"2"       validate:"required,gte=1,lte=32"`
}
