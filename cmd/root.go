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

package cmd

import (
	"fmt"
	cloneCmd "github.com/skyscanner/turbolift/cmd/clone"
	commitCmd "github.com/skyscanner/turbolift/cmd/commit"
	createPrsCmd "github.com/skyscanner/turbolift/cmd/create_prs"
	"github.com/skyscanner/turbolift/cmd/flags"
	foreachCmd "github.com/skyscanner/turbolift/cmd/foreach"
	initCmd "github.com/skyscanner/turbolift/cmd/init"
	"github.com/spf13/cobra"
	"log"
)

var version = "version-dev"
var commit = "commit-dev"
var date = "date-dev"

var rootCmd = &cobra.Command{
	Use:     "turbolift",
	Short:   "Turbolift",
	Long:    `Mass refactoring tool for repositories in GitHub`,
	Version: fmt.Sprintf("%s (%s, built %s)", version, commit, date),
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flags.Verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(cloneCmd.NewCloneCmd())
	rootCmd.AddCommand(commitCmd.NewCommitCmd())
	rootCmd.AddCommand(createPrsCmd.NewCreatePRsCmd())
	rootCmd.AddCommand(initCmd.NewInitCmd())
	rootCmd.AddCommand(foreachCmd.NewForeachCmd())
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
