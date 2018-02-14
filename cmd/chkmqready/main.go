/*
© Copyright IBM Corporation 2017, 2018

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// chkmqready checks that MQ is ready for work, by checking if the MQ listener port is available
package main

import (
	"net"
	"os"

	"github.com/ibm-messaging/mq-container/internal/ready"
)

func main() {
	// Check if runmqserver has indicated that it's finished configuration
	r, err := ready.Check()
	if !r || err != nil {
		os.Exit(1)
	}
	// Check if the queue manager has a running listener
	conn, err := net.Dial("tcp", "127.0.0.1:1414")
	if err != nil {
		os.Exit(1)
	}
	conn.Close()
}
