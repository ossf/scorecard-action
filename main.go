// Copyright OpenSSF Authors
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
	"github.com/ossf/scorecard-action/entrypoint"
	"github.com/ossf/scorecard-action/options"
)

var opts = &options.Options{}

func main() {
	err := entrypoint.Run(opts)
	if err != nil {
		// TODO: Don't panic!
		panic(err)
	}
}
