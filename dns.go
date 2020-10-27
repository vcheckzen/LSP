package main

import (
	"net"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

var (
	clients = &sync.Pool{New: func() interface{} {
		return &dns.Client{Timeout: time.Second * 3}
	}}
	localIp = net.ParseIP("127.0.0.1")
)

func listenDns() {
	if dnsUpstreams.isEmpty() {
		log.Warning("No dns upstream, will not start dns server.")
		return
	}

	var handler dns.HandlerFunc = relayDns
	log.Fatal("Can not start dns server. ",
		dns.ListenAndServe(":53", "udp", handler),
	)
}

func redirectToLocal(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Answer = []dns.RR{
		&dns.A{
			Hdr: dns.RR_Header{
				Name:   r.Question[0].Name,
				Rrtype: dns.TypeA,
				Class:  dns.ClassINET,
				Ttl:    60,
			},
			A: localIp,
		},
	}
	if err := w.WriteMsg(m); err != nil {
		log.Warning("Can not replay dns client with local ip. ", err)
	}
}

func relayDns(w dns.ResponseWriter, r *dns.Msg) {
	host := strings.TrimRight(r.Question[0].Name, ".")
	if _, ok := hostResolver.get(host); ok {
		redirectToLocal(w, r)
		return
	}

	client := clients.Get().(*dns.Client)
	defer clients.Put(client)

	upstreams := dnsUpstreams.keys()
	for _, upstream := range upstreams {
		splits := strings.Split(upstream.(string), "://")
		client.Net = splits[0]
		if rst, _, err := client.Exchange(r, splits[1]); err == nil {
			if err := w.WriteMsg(rst); err == nil {
				break
			} else {
				log.Warning("Dns exchange failed. ", err)
			}
		} else {
			log.Warning("Dns exchange failed. ", err)
		}
	}
}
