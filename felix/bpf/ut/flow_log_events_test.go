// Copyright (c) 2024-2025 Tigera, Inc. All rights reserved.
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

package ut_test

import (
	"testing"

	"github.com/google/gopacket/layers"
	. "github.com/onsi/gomega"

	"github.com/projectcalico/calico/felix/bpf/events"
	"github.com/projectcalico/calico/felix/bpf/perf"
	"github.com/projectcalico/calico/felix/bpf/routes"
)

func TestFlowLogV6Events(t *testing.T) {
	RegisterTestingT(t)
	hostIP = node1ipV6

	_, ip6hdr, l4, _, pktBytes, err := testPacket(6, nil, nil, nil, nil)
	Expect(err).NotTo(HaveOccurred())
	ipv6 := ip6hdr.(*layers.IPv6)
	udp := l4.(*layers.UDP)

	perfEvents, err := perf.New(perfMap, 1<<20)
	Expect(err).NotTo(HaveOccurred())
	defer perfEvents.Close()

	rtKey := routes.NewKeyV6(srcV6CIDR).AsBytes()
	rtVal := routes.NewValueV6WithIfIndex(routes.FlagsLocalWorkload, 1).AsBytes()
	err = rtMapV6.Update(rtKey, rtVal)
	Expect(err).NotTo(HaveOccurred())

	skbMark = 0
	runBpfTest(t, "calico_from_workload_ep", rulesDefaultAllow, func(bpfrun bpfProgRunFn) {
		res, err := bpfrun(pktBytes)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Retval).To(Equal(resTC_ACT_UNSPEC))
	}, withIPv6())

	rawEvent, err := perfEvents.Next()
	Expect(err).NotTo(HaveOccurred())
	e, err := events.ParseEvent(rawEvent)
	Expect(err).NotTo(HaveOccurred())
	evnt := events.ParsePolicyVerdict(e.Data(), true)
	Expect(evnt.SrcAddr).To(Equal(ipv6.SrcIP))
	Expect(evnt.DstAddr).To(Equal(ipv6.DstIP))
	Expect(evnt.SrcPort).To(Equal(uint16(udp.SrcPort)))
	Expect(evnt.DstPort).To(Equal(uint16(udp.DstPort)))
	Expect(evnt.IPProto).To(Equal(uint8(ipv6.NextHeader)))
}

func TestFlowLogEvents(t *testing.T) {
	RegisterTestingT(t)
	hostIP = node1ip

	_, iphdr, l4, _, pktBytes, err := testPacket(4, nil, nil, nil, nil)
	Expect(err).NotTo(HaveOccurred())
	ipv4 := iphdr.(*layers.IPv4)
	udp := l4.(*layers.UDP)

	perfEvents, err := perf.New(perfMap, 1<<20)
	Expect(err).NotTo(HaveOccurred())
	defer perfEvents.Close()

	rtKey := routes.NewKey(srcV4CIDR).AsBytes()
	rtVal := routes.NewValueWithIfIndex(routes.FlagsLocalWorkload, 1).AsBytes()
	err = rtMap.Update(rtKey, rtVal)
	Expect(err).NotTo(HaveOccurred())

	skbMark = 0
	runBpfTest(t, "calico_from_workload_ep", rulesDefaultAllow, func(bpfrun bpfProgRunFn) {
		res, err := bpfrun(pktBytes)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Retval).To(Equal(resTC_ACT_UNSPEC))
	})

	rawEvent, err := perfEvents.Next()
	Expect(err).NotTo(HaveOccurred())
	e, err := events.ParseEvent(rawEvent)
	Expect(err).NotTo(HaveOccurred())
	evnt := events.ParsePolicyVerdict(e.Data(), false)
	Expect(evnt.SrcAddr).To(Equal(ipv4.SrcIP))
	Expect(evnt.DstAddr).To(Equal(ipv4.DstIP))
	Expect(evnt.SrcPort).To(Equal(uint16(udp.SrcPort)))
	Expect(evnt.DstPort).To(Equal(uint16(udp.DstPort)))
	Expect(evnt.IPProto).To(Equal(uint8(ipv4.Protocol)))

	// This packet follows the established flow so should not generate any event ...
	skbMark = 0
	runBpfTest(t, "calico_from_workload_ep", rulesDefaultAllow, func(bpfrun bpfProgRunFn) {
		res, err := bpfrun(pktBytes)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Retval).To(Equal(resTC_ACT_UNSPEC))
	})

	// ... however, this is a packet for a different new flow so that should generate the
	// next event. It must come right after the event generated by the first packet of the
	// first flow.
	udp.SrcPort = 60606
	_, _, _, _, pktBytes, err = testPacket(4, nil, nil, udp, nil)
	Expect(err).NotTo(HaveOccurred())

	skbMark = 0
	runBpfTest(t, "calico_from_workload_ep", rulesDefaultAllow, func(bpfrun bpfProgRunFn) {
		res, err := bpfrun(pktBytes)
		Expect(err).NotTo(HaveOccurred())
		Expect(res.Retval).To(Equal(resTC_ACT_UNSPEC))
	})

	rawEvent, err = perfEvents.Next()
	Expect(err).NotTo(HaveOccurred())
	e, err = events.ParseEvent(rawEvent)
	Expect(err).NotTo(HaveOccurred())
	evnt = events.ParsePolicyVerdict(e.Data(), false)
	Expect(evnt.SrcAddr).To(Equal(ipv4.SrcIP))
	Expect(evnt.DstAddr).To(Equal(ipv4.DstIP))
	Expect(evnt.SrcPort).To(Equal(uint16(udp.SrcPort)))
	Expect(evnt.DstPort).To(Equal(uint16(udp.DstPort)))
	Expect(evnt.IPProto).To(Equal(uint8(ipv4.Protocol)))
}
