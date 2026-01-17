package api_tests

import (
	"encoding/json"
	"net/http"

	. "github.com/onsi/gomega"
)

func decodeResponseBody(resp *http.Response) map[string]interface{} {
	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	Expect(err).NotTo(HaveOccurred())
	return body
}

func expectBusinessCode(resp *http.Response, expected int) map[string]interface{} {
	body := decodeResponseBody(resp)
	Expect(body["code"]).To(BeEquivalentTo(expected))
	return body
}

func expectBusinessCodeNotEqual(resp *http.Response, unexpected int) map[string]interface{} {
	body := decodeResponseBody(resp)
	Expect(body["code"]).NotTo(BeEquivalentTo(unexpected))
	return body
}
