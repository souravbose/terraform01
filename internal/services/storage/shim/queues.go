// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shim

import (
	"context"

	"github.com/tombuildsstuff/giovanni/storage/2020-08-04/queue/queues"
)

type StorageQueuesWrapper interface {
	Create(ctx context.Context, resourceGroup, accountName, queueName string, metaData map[string]string) error
	Delete(ctx context.Context, resourceGroup, accountName, queueName string) error
	Exists(ctx context.Context, resourceGroup, accountName, queueName string) (*bool, error)
	Get(ctx context.Context, resourceGroup, accountName, queueName string) (*StorageQueueProperties, error)
	GetServiceProperties(ctx context.Context, resourceGroup, accountName string) (*queues.StorageServiceProperties, error)
	UpdateMetaData(ctx context.Context, resourceGroup, accountName, queueName string, metaData map[string]string) error
	UpdateServiceProperties(ctx context.Context, resourceGroup, accountName string, properties queues.StorageServiceProperties) error
}

type StorageQueueProperties struct {
	MetaData map[string]string
}
