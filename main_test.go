// Copyright 2026 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ossf/scorecard-action/internal/publication"
	"github.com/ossf/scorecard-action/options"
	"github.com/ossf/scorecard-action/signing"
)

var errTest = errors.New("test error")

func Test_getFailOnPublishError(t *testing.T) {
	tests := []struct {
		value   string
		want    bool
		wantErr bool
	}{
		{value: "", want: false},
		{value: "false", want: false},
		{value: "true", want: true},
		{value: "yes", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Setenv(options.EnvInputFailOnPublishError, tt.value)
			got, err := getFailOnPublishError()
			if (err != nil) != tt.wantErr {
				t.Fatalf("getFailOnPublishError() error: %v, wantErr: %t", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("getFailOnPublishError(): got %t, wanted %t", got, tt.want)
			}
		})
	}
}

func Test_reportPublishError(t *testing.T) {
	t.Parallel()
	rejected := publication.Result{
		Attempted: true,
		Err:       &signing.PublishError{StatusCode: http.StatusBadRequest, Err: errTest},
	}
	unavailable := publication.Result{
		Attempted: true,
		Err:       &signing.PublishError{StatusCode: http.StatusServiceUnavailable, Err: errTest},
	}
	tests := []struct {
		name               string
		pub                publication.Result
		failOnPublishError bool
		wantFail           bool
	}{
		{name: "rejected, strict", pub: rejected, failOnPublishError: true, wantFail: true},
		{name: "rejected, default", pub: rejected, failOnPublishError: false, wantFail: false},
		{name: "unavailable, strict", pub: unavailable, failOnPublishError: true, wantFail: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := reportPublishError(tt.pub, tt.failOnPublishError); got != tt.wantFail {
				t.Errorf("reportPublishError(): got %t, wanted %t", got, tt.wantFail)
			}
		})
	}
}
