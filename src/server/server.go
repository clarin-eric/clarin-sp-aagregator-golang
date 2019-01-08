package server

import (
	"net/http"
	"strconv"
	"clarin/shib-aagregator/src/logger"
)

func StartServerAndblock(logPtr *logger.Logger, port int, sp_entity_id string, simulate bool, aagregator_url, aagregator_path string) {
	entity_id_map := map[string]string{
		"default": "https://test-sp.clarin.eu",
		"sp1": "https://test-sp1.clarin.eu",
		"sp2": "https://test-sp2.clarin.eu",
	}

	h := Handler{
		logPtr: logPtr,
		entity_id_map: entity_id_map,
		simulate: simulate,
		aggregator_path: aagregator_path,
		aggregator_url: aagregator_url,
	}

	//Register handlers
	http.HandleFunc("/", h.handler)
	//Start server
	logPtr.Info("Log level: %s", logPtr.GetLevelAsString())
	logPtr.Info("Simulation: %t", simulate)
	logPtr.Info("Starting http server on port: %d", port)

	err := http.ListenAndServe(":"+strconv.Itoa(port), nil)
	if err != nil {
		logPtr.Error("Failed to start server: %v", err)
	}
}



