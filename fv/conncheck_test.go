// Copyright (c) 2020 Tigera, Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build fvtests

package fv_test

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/projectcalico/libcalico-go/lib/apiconfig"

	"github.com/projectcalico/felix/fv/connectivity"
	"github.com/projectcalico/felix/fv/infrastructure"
	"github.com/projectcalico/felix/fv/tcpdump"
	"github.com/projectcalico/felix/fv/workload"
)

// These tests verify that test-connection and test-workload work properly across all the different protocols.
var _ = describeConnCheckTests("tcp")
var _ = describeConnCheckTests("sctp")
var _ = describeConnCheckTests("udp")
var _ = describeConnCheckTests("udp-recvmsg")
var _ = describeConnCheckTests("udp-noconn")

func describeConnCheckTests(protocol string) bool {
	return infrastructure.DatastoreDescribe("Connectivity checker self tests: "+protocol,
		[]apiconfig.DatastoreType{apiconfig.EtcdV3}, // Skipping k8s since these tests don't rely on the datastore.
		func(getInfra infrastructure.InfraFactory) {

			var (
				infra   infrastructure.DatastoreInfra
				felixes []*infrastructure.Felix
				hostW   [2]*workload.Workload
				cc      *connectivity.Checker
			)

			BeforeEach(func() {
				infra = getInfra()
				felixes, _ = infrastructure.StartNNodeTopology(2, infrastructure.DefaultTopologyOptions(), infra)

				// Create host-networked "workloads", one on each "host".
				for ii := range felixes {
					// Workload doesn't understand the extra connectivity types that test-connection tries.
					wlProto := protocol
					if strings.Contains(protocol, "-") {
						wlProto = strings.Split(protocol, "-")[0]
					}
					hostW[ii] = workload.Run(felixes[ii], fmt.Sprintf("host%d", ii), "", felixes[ii].IP, "8055", wlProto)
				}

				cc = &connectivity.Checker{}
				cc.Protocol = protocol
			})

			AfterEach(func() {
				for _, wl := range hostW {
					wl.Stop()
				}
				for _, felix := range felixes {
					felix.Stop()
				}

				if CurrentGinkgoTestDescription().Failed {
					infra.DumpErrorData()
				}
				infra.Stop()
			})

			It("should have host-to-host on right port only", func() {
				cc.ExpectSome(felixes[0], hostW[1])
				cc.ExpectNone(felixes[0], hostW[1], 8066)
				cc.CheckConnectivity()
			})

			if protocol == "udp" {
				Describe("with tc configured to drop 5% of packets", func() {
					BeforeEach(func() {
						// Make sure we have connectivity before we start packet loss measurements.
						cc.ExpectSome(felixes[0], hostW[1])
						cc.CheckConnectivity()
						cc.ResetExpectations()

						felixes[0].Exec("tc", "qdisc", "add", "dev", "eth0", "root", "netem", "loss", "5%")
					})

					It("and a 1% threshold, should see packet loss", func() {
						failed := false
						cc.OnFail = func(msg string) {
							log.WithField("msg", msg).Info("Connectivity checker failed (as expected)")
							failed = true
						}
						cc.ExpectLoss(felixes[0], hostW[1], 2*time.Second, 1, -1)
						cc.CheckConnectivityPacketLoss()

						Expect(failed).To(BeTrue(), "Expected the connection checker to detect packet loss")
					})

					It("and a 20% threshold, should tolerate packet loss", func() {
						cc.ExpectLoss(felixes[0], hostW[1], 2*time.Second, 20, -1)
						cc.CheckConnectivityPacketLoss()
					})

					It("with tcpdump", func() {
						tcpd := tcpdump.Attach(felixes[0].Container, "eth0")
						tcpd.SetLogEnabled(true)
						tcpd.AddMatcher("UDP", regexp.MustCompile(`.*UDP.*`))
						tcpd.Start()
						cc.ExpectLoss(felixes[0], hostW[1], 2*time.Second, 20, -1)
						cc.CheckConnectivityPacketLoss()
						Eventually(func() int { return tcpd.MatchCount("UDP") }).Should(BeNumerically(">", 0))
					})
				})
			}
		},
	)
}
