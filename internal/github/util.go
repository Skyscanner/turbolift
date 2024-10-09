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

// util contains helper functions needed for both real and fake implementations of the GitHub interface

package github

import "encoding/json"

type ViewerPermission struct {
	ViewerPermission string `json:"viewerPermission"`
}

// IsPushable checks the output of the viewerPermission API query
// and returns true if the user has write, maintain or admin permissions

func IsPushable(viewerPermissionOutput string) (bool, error) {
	var vp ViewerPermission

	if err := json.Unmarshal([]byte(viewerPermissionOutput), &vp); err != nil {
		return false, err
	}

	switch vp.ViewerPermission {
	case "WRITE", "MAINTAIN", "ADMIN":
		return true, nil
	default:
		return false, nil
	}
}
