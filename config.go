package main

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strings"

	logger "github.com/sirupsen/logrus"
)

type hostInfo struct {
	addr string
	sn   string
}

const (
	dnsServerPath = "data/dns"
	hostsPath     = "data/hosts"
	caCertPath    = "data/ca.crt"
	caKeyPath     = "data/ca.key"
	defaultSN     = "www.apple.com"
	apiFilePath   = "static"
	apiDomain     = "lsp.com"
	apiAddr       = "127.0.0.1:3080"
)

var (
	log          = logger.New()
	dnsUpstreams = newSyncMap()
	hostResolver = newSyncMap()
)

func init() {
	parserProxy(hostsPath, true, parseHost)
	parserProxy(dnsServerPath, true, parseDnsServer)
	refreshDNS()
}

func parserProxy(path string, skipEmpty bool, lineParser func(line string)) {
	file, err := os.Open(path)
	if err != nil {
		log.Fatalf("Can not open %s file. %s", path, err)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if skipEmpty &&
			(len(line) == 0 || line[0] == '#') {
			continue
		}
		lineParser(line)
	}

	if err := file.Close(); err != nil {
		log.Fatal("Can not close %s file. %s", path, err)
	}
}

func parseHost(line string) {
	fields := strings.Fields(line)
	info := &hostInfo{addr: fields[0] + ":https", sn: defaultSN}
	if len(fields) == 3 {
		info.sn = fields[2]
	}
	hostResolver.add(fields[1], info)
}

func parseDnsServer(line string) {
	dnsUpstreams.add(line, true)
}

func saveAndRefresh(path, lines string, parser func(line string)) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		log.Warning("Can not open file. ", err)
		return
	}
	defer func() { _ = f.Close() }()

	w := bufio.NewWriter(f)

	for _, line := range strings.Split(lines, "\n") {
		line := strings.TrimSpace(line)

		if _, err := w.WriteString(line + "\n"); err != nil {
			log.Warning("Can not write to file. ", err)
		}

		if len(line) != 0 && line[0] != '#' {
			parser(line)
		}
	}

	_ = w.Flush()
}

func refreshConfig(dns, hosts string) {
	if dns != "" {
		dnsUpstreams.clear()
		saveAndRefresh(dnsServerPath, dns, parseDnsServer)
	}

	if hosts != "" {
		hostResolver.clear()
		saveAndRefresh(hostsPath, hosts, parseHost)
	}

	refreshDNS()
}

func refreshDNS() {
	// relay api
	hostResolver.add(apiDomain, &hostInfo{addr: apiAddr})

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd.exe", "/C", "ipconfig", "/flushdns")
	case "darwin":
		cmd = exec.Command("/bin/bash", "killall", "-HUP", "mDNSResponder")
	}
	if cmd != nil {
		stdout, _ := cmd.CombinedOutput()
		log.Info(string(stdout))
	}
}
