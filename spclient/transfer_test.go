//go:build test_unit

package spclient_test

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/devgianlu/go-librespot/spclient"
)

func (suite *RequestSuite) TestTransferFromActiveNamesThisDeviceAtBothEnds() {
	suite.Require().NoError(suite.spclient.TransferFromActive(suite.T().Context()))

	got := suite.requests()
	suite.Require().Len(got, 1)
	suite.Equal("POST", got[0].method)
	suite.Equal("/connect-state/v1/connect/transfer/from/device-id/to/device-id", got[0].path)
	suite.JSONEq(`{"transfer_options":{"restore_paused":"resume"}}`, string(got[0].body))
}

// A transfer is not idempotent: one resent after Spotify acted on it would
// arrive while this device is already active.
func (suite *RequestSuite) TestTransferFromActiveIsSentOnce() {
	suite.handler = func(_ int, w http.ResponseWriter) { w.WriteHeader(http.StatusBadGateway) }

	// Bounded, so a retrying request fails here rather than at the suite's timeout.
	ctx, cancel := context.WithTimeout(suite.T().Context(), 3*time.Second)
	defer cancel()
	err := suite.spclient.TransferFromActive(ctx)

	var status *spclient.StatusError
	suite.Require().True(errors.As(err, &status))
	suite.Equal(http.StatusBadGateway, status.StatusCode)
	suite.Len(suite.requests(), 1)
}
