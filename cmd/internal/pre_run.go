/*
 * Copyright 2021 Skyscanner Limited.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 * http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package internal

import (
	"github.com/skyscanner/turbolift/cmd/flags"
	"github.com/skyscanner/turbolift/internal/logging"
	"github.com/spf13/cobra"
)

// PreRun is a common pre-run function for all commands
func PreRun(cmd *cobra.Command, args []string) {
	logger := logging.NewLogger(cmd)

	if flags.DryRun {
		logger.Infof("turbolift %s is in dry-run mode\n", cmd.Use)
	}
}
