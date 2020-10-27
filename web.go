package main

import (
	"strings"

	"encoding/json"
	"net/http"
)

type config struct {
	Dns   string `json:"dns"`
	Hosts string `json:"hosts"`
}

func listenApi() {
	var (
		handleGetConfig  http.HandlerFunc = getConfig
		handleSaveConfig http.HandlerFunc = saveConfig
	)

	http.Handle("/", http.FileServer(http.Dir(apiFilePath)))
	http.Handle("/get-config/", handleGetConfig)
	http.Handle("/save-config/", handleSaveConfig)

	log.Fatal(
		"Can not start api server. ",
		http.ListenAndServeTLS(apiAddr, caCertPath, caKeyPath, nil),
	)
}

func saveConfig(w http.ResponseWriter, r *http.Request) {
	var cfg config
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		log.Warning("Can not decode request body. ", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	refreshConfig(cfg.Dns, cfg.Hosts)
	w.WriteHeader(http.StatusOK)
}

func getConfig(w http.ResponseWriter, _ *http.Request) {
	cfg := new(config)
	parserProxy(dnsServerPath, false, func(line string) {
		cfg.Dns += line + "\n"
	})
	parserProxy(hostsPath, false, func(line string) {
		cfg.Hosts += line + "\n"
	})
	cfg.Dns = strings.TrimSpace(cfg.Dns)
	cfg.Hosts = strings.TrimSpace(cfg.Hosts)

	var err error
	if js, err := json.Marshal(&cfg); err == nil {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(js); err == nil {
			return
		}
	}
	log.Warning("Send config failed. ", err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
