package server

import (
	"fmt"
	"net/http"
	"clarin/shib-aagregator/src/logger"
)

type Handler struct {
	entity_id_map map[string]string //map of application id's to entity ids
	simulate bool
	aggregator_url string
	aggregator_path string
	logPtr *logger.Logger
}

func (h *Handler) handler(w http.ResponseWriter, r *http.Request) {
	//Fetch return location from query string
	redirectLocations := h.getValues("return", r.URL.Query())
	if len(redirectLocations) < 0 {
		h.httpError(w, "No return parameter found in query parameters.\n")
		http.Redirect(w, r, redirectLocations[0], 302)
		return
	}

	//Fetch application id from request headers
	shib_application_id := h.getFirstValue("Shib-Application-Id", r.Header)
	ok := false
	sp_entity_id := ""
	if sp_entity_id, ok = h.entity_id_map[shib_application_id]; !ok {
		h.httpError(w, "Failed to get application id (Shib-Application-Id) header.\n")
		http.Redirect(w, r, redirectLocations[0], 302)
		return
	}

	h.logPtr.Debug("Mapped application id: %s to entity id: %s", shib_application_id, sp_entity_id)

	//Fetch assertion count number from request headers
	shib_assertion_count := h.getFirstValue("Shib-Assertion-Count", r.Header)
	if len(shib_assertion_count) < 0 {
		h.httpError(w, "Failed to get assertion count (Shib-Assertion-Count) headers. Double check if `exportAssertion=\"true\"` is configured.\n")
		http.Redirect(w, r, redirectLocations[0], 302)
		return
	}

	//Fetch assertion url number from request headers
	shib_assertion_url := h.getFirstValue(fmt.Sprintf("Shib-Assertion-%s", shib_assertion_count), r.Header)
	if len(shib_assertion_url) <= 0 {
		h.httpError(w, "Failed to get assertion url (Shib-Assertion-%s) from headers. Double check if `exportAssertion=\"true\"` is configured.\n", shib_assertion_count)
		http.Redirect(w, r, redirectLocations[0], 302)
		return
	}

	/*
	//TODO: make this step optional
	reg, err := regexp.Compile("https://(.+)/Shibboleth.sso/(.+)")
	if err != nil {
		h.httpError(w, "Failed to parse regular expression. Error: %s\n", err)
		http.Redirect(w, r, redirectLocations[0], 302)
		return
	}

	url_parts := reg.FindStringSubmatch(shib_assertion_url)
	if url_parts == nil {
		h.httpError(w, "Failed to parse assertion url.\n", err)
		http.Redirect(w, r, redirectLocations[0], 302)
		return
	}
	//TODO: check length of url_parts
	//TODO: make sp_host configurable
	sp_host := "proxy"
	processed_shib_assertion_url := fmt.Sprintf("https://%s/Shibboleth.sso/%s", sp_host, url_parts[2])

	h.logPtr.Info("shib_assertion_url=%s, processed_shib_assertion_url=%s\n", shib_assertion_url, processed_shib_assertion_url)
	*/

	//Get SAML assertions from SP
	attrInfo, err := h.getAttributeAssertions(shib_assertion_url, sp_entity_id)
	if err != nil {
		h.httpError(w, "Failed to get SAML assertion from SP. Assertion url=%s, entity id=%s. Error: %v\n", shib_assertion_url, sp_entity_id, err)
		http.Redirect(w, r, redirectLocations[0], 302)
		return
	}

	//Submit information to aagregator
	err = h.sendInfo(attrInfo)
	if err != nil {
		h.httpError(w, "Failed to send attributes. Error: %v", err)
		http.Redirect(w, r, redirectLocations[0], 302)
		return
	}

	//No errors, so redirect to return location
	http.Redirect(w, r, redirectLocations[0], 302)
}

func (h *Handler) httpError(w http.ResponseWriter, _fmt string, args ...interface{}) {
	msg := fmt.Sprintf(_fmt, args...)
	//Log issue to stdout
	h.logPtr.Error(msg)
	//Send error to client
	//w.WriteHeader(http.StatusInternalServerError)
	//w.Write([]byte(fmt.Sprintf("500 - %s", msg)))
}

func (h *Handler) getFirstValue(key string, m map[string][]string) (string) {
	values := h.getValues(key, m)
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func (h *Handler) getValues(key string, m map[string][]string) ([]string) {
	if value, ok := m[key]; ok {
		return value
	}
	return nil
}