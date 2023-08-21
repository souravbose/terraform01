// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package queues

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/validation"
	"github.com/tombuildsstuff/giovanni/storage/internal/endpoints"
	"github.com/tombuildsstuff/giovanni/storage/internal/metadata"
)

// Create creates the specified Queue within the specified Storage Account
func (client Client) Create(ctx context.Context, accountName, queueName string, metaData map[string]string) (result autorest.Response, err error) {
	if accountName == "" {
		return result, validation.NewError("queues.Client", "Create", "`accountName` cannot be an empty string.")
	}
	if queueName == "" {
		return result, validation.NewError("queues.Client", "Create", "`queueName` cannot be an empty string.")
	}
	if strings.ToLower(queueName) != queueName {
		return result, validation.NewError("queues.Client", "Create", "`queueName` must be a lower-cased string.")
	}
	if err := metadata.Validate(metaData); err != nil {
		return result, validation.NewError("queues.Client", "Create", fmt.Sprintf("`metadata` is not valid: %s.", err))
	}

	req, err := client.CreatePreparer(ctx, accountName, queueName, metaData)
	if err != nil {
		err = autorest.NewErrorWithError(err, "queues.Client", "Create", nil, "Failure preparing request")
		return
	}

	resp, err := client.CreateSender(req)
	if err != nil {
		result = autorest.Response{Response: resp}
		err = autorest.NewErrorWithError(err, "queues.Client", "Create", resp, "Failure sending request")
		return
	}

	result, err = client.CreateResponder(resp)
	if err != nil {
		err = autorest.NewErrorWithError(err, "queues.Client", "Create", resp, "Failure responding to request")
		return
	}

	return
}

// CreatePreparer prepares the Create request.
func (client Client) CreatePreparer(ctx context.Context, accountName string, queueName string, metaData map[string]string) (*http.Request, error) {
	pathParameters := map[string]interface{}{
		"queueName": autorest.Encode("path", queueName),
	}

	headers := map[string]interface{}{
		"x-ms-version": APIVersion,
	}

	headers = metadata.SetIntoHeaders(headers, metaData)

	preparer := autorest.CreatePreparer(
		autorest.AsContentType("application/xml; charset=utf-8"),
		autorest.AsPut(),
		autorest.WithBaseURL(endpoints.GetQueueEndpoint(client.BaseURI, accountName)),
		autorest.WithPathParameters("/{queueName}", pathParameters),
		autorest.WithHeaders(headers))
	return preparer.Prepare((&http.Request{}).WithContext(ctx))
}

// CreateSender sends the Create request. The method will close the
// http.Response Body if it receives an error.
func (client Client) CreateSender(req *http.Request) (*http.Response, error) {
	return autorest.SendWithSender(client, req,
		azure.DoRetryWithRegistration(client.Client))
}

// CreateResponder handles the response to the Create request. The method always
// closes the http.Response Body.
func (client Client) CreateResponder(resp *http.Response) (result autorest.Response, err error) {
	err = autorest.Respond(
		resp,
		client.ByInspecting(),
		azure.WithErrorUnlessStatusCode(http.StatusCreated),
		autorest.ByClosing())
	result = autorest.Response{Response: resp}

	return
}
