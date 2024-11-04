/*
 * Copyright 2021 Skyscanner Limited.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * https://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestuserHasPushPermissionReturnsTrueForAllCases(t *testing.T) {
	testCases := []string{
		`{"viewerPermission":"WRITE"}`,
		`{"viewerPermission":"MAINTAIN"}`,
		`{"viewerPermission":"ADMIN"}`,
	}

	for _, testCase := range testCases {
		pushable, err := userHasPushPermission(testCase)
		assert.NoError(t, err)
		assert.True(t, pushable)
	}
}

func TestuserHasPushPermissionReturnsFalseForUnknownPermission(t *testing.T) {
	testCases := []string{
		`{"viewerPermission":"UNKNOWN"}`,
		`{"viewerPermission":"READ"}`,
		`{"viewerPermission":"RANDOM"}`,
		`{"viewerPermission":""}`,
		`{"anotherParam":"WRITE"}`, // valid JSON but not the expected format
	}

	for _, testCase := range testCases {
		pushable, err := userHasPushPermission(testCase)
		assert.NoError(t, err)
		assert.False(t, pushable)
	}
}

func TestuserHasPushPermissionReturnsErrorForInvalidJSON(t *testing.T) {
	testCases := []string{
		`{"viewerPermission":"WRITE"`, // invalid JSON
		`viewerPermission: WRITE`,     // invalid JSON
		`{"viewerPermission": WRITE}`, // invalid JSON
	}

	for _, testCase := range testCases {
		_, err := userHasPushPermission(testCase)
		assert.Error(t, err)
	}
}
