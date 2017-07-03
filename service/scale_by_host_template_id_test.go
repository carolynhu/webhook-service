package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	v1client "github.com/rancher/go-rancher/client"
	"github.com/rancher/go-rancher/v2"
	"github.com/rancher/webhook-service/drivers"
	"github.com/rancher/webhook-service/model"
)

func TestWebhookCreateAndExecuteWithHostTemplateID(t *testing.T) {
	// Test creating a webhook
	constructURL := fmt.Sprintf("%s/v1-webhooks/receivers?projectId=1a5", server.URL)
	jsonStr := []byte(`{"driver": "scaleByHostTemplateID", "name": "wh-name",
	"scaleByHostTemplateIDConfig": {"action": "up",	"amount": 1, "hostTemplateId": "1ht1", "min": 1, "max": 4, "deleteOption": "mostRecent"}}`)
	request, err := http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler := HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("StatusCode %d means ConstructPayloadTest failed", response.Code)
	}
	resp, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	wh := &model.Webhook{}
	err = json.Unmarshal(resp, wh)
	if err != nil {
		t.Fatal(err)
	}
	if wh.Name != "wh-name" || wh.Driver != "scaleByHostTemplateID" || wh.Id != "1" || wh.URL == "" ||
		wh.ScaleByHostTemplateIDConfig.Action != "up" || wh.ScaleByHostTemplateIDConfig.Amount != 1 || wh.ScaleByHostTemplateIDConfig.Min != 1 ||
		wh.ScaleByHostTemplateIDConfig.Max != 4 || wh.ScaleByHostTemplateIDConfig.Type != "scaleByHostTemplateID" || wh.ScaleByHostTemplateIDConfig.HostTemplateID != "1ht1" {
		t.Fatalf("Unexpected webhook: %#v", wh)
	}
	if !strings.HasSuffix(wh.Links["self"], "/v1-webhooks/receivers/1?projectId=1a5") {
		t.Fatalf("Bad self URL: %v", wh.Links["self"])
	}

	// Test getting the created webhook by id
	byID := fmt.Sprintf("%s/v1-webhooks/receivers/1?projectId=1a5", server.URL)
	request, err = http.NewRequest("GET", byID, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("StatusCode %d means get failed", response.Code)
	}
	resp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	wh = &model.Webhook{}
	err = json.Unmarshal(resp, wh)
	if err != nil {
		t.Fatal(err)
	}
	if wh.Name != "wh-name" || wh.Driver != "scaleByHostTemplateID" || wh.Id != "1" || wh.URL == "" ||
		wh.ScaleByHostTemplateIDConfig.Action != "up" || wh.ScaleByHostTemplateIDConfig.Amount != 1 || wh.ScaleByHostTemplateIDConfig.Min != 1 ||
		wh.ScaleByHostTemplateIDConfig.Max != 4 || wh.ScaleByHostTemplateIDConfig.Type != "scaleByHostTemplateID" || wh.ScaleByHostTemplateIDConfig.HostTemplateID != "1ht1" {
		t.Fatalf("Unexpected webhook: %#v", wh)
	}
	if !strings.HasSuffix(wh.Links["self"], "/v1-webhooks/receivers/1?projectId=1a5") {
		t.Fatalf("Bad self URL: %v", wh.Links["self"])
	}

	// Test executing the webhook
	url := wh.URL
	requestExecute, err := http.NewRequest("POST", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	response = httptest.NewRecorder()
	handler = HandleError(schemas, r.Execute)
	handler.ServeHTTP(response, requestExecute)
	if response.Code != 200 {
		t.Errorf("StatusCode %d means execute failed", response.Code)
	}

	//List webhooks
	requestList, err := http.NewRequest("GET", constructURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	requestList.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, requestList)
	if response.Code != 200 {
		t.Fatalf("StatusCode %d means get failed", response.Code)
	}
	resp, err = ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	whCollection := &model.WebhookCollection{}
	err = json.Unmarshal(resp, whCollection)
	if err != nil {
		t.Fatal(err)
	}
	if len(whCollection.Data) != 1 {
		t.Fatal("Added webhook not listed")
	}
	wh = &whCollection.Data[0]
	if wh.Name != "wh-name" || wh.Driver != "scaleByHostTemplateID" || wh.Id != "1" || wh.URL == "" ||
		wh.ScaleByHostTemplateIDConfig.Action != "up" || wh.ScaleByHostTemplateIDConfig.Amount != 1 || wh.ScaleByHostTemplateIDConfig.Min != 1 ||
		wh.ScaleByHostTemplateIDConfig.Max != 4 || wh.ScaleByHostTemplateIDConfig.Type != "scaleByHostTemplateID" || wh.ScaleByHostTemplateIDConfig.HostTemplateID != "1ht1" {
		t.Fatalf("Unexpected webhook: %#v", wh)
	}
	if !strings.HasSuffix(wh.Links["self"], "/v1-webhooks/receivers/1?projectId=1a5") {
		t.Fatalf("Bad self URL: %v", wh.Links["self"])
	}

	//Delete
	request, err = http.NewRequest("DELETE", byID, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != 204 {
		t.Fatalf("StatusCode %d means delete failed", response.Code)
	}
}

func TestWebhookCreateInvalidMinMaxActionWithHostTemplateID(t *testing.T) {
	constructURL := fmt.Sprintf("%s/v1-webhooks/receivers?projectId=1a5", server.URL)
	jsonStr := []byte(`{"driver":"scaleByHostTemplateID","name":"wh-name",
		"scaleByHostTemplateIDConfig": {"action": "up", "amount": 1, "hostTemplateId": "1ht1",  "min": -1, "max": 4, "deleteOption": "mostRecent"}}`)
	request, err := http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	handler := HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code == 200 {
		t.Fatalf("Invalid min")
	}

	jsonStr = []byte(`{"driver":"scaleByHostTemplateID","name":"wh-name",
		"scaleByHostTemplateIDConfig": {"action": "up", "amount": 1, "hostTemplateId": "1ht1",  "min": 1, "max": -4, "deleteOption": "mostRecent"}}`)
	request, err = http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler = HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code == 200 {
		t.Fatalf("Invalid max")
	}

	jsonStr = []byte(`{"driver":"scaleByHostTemplateID","name":"wh-name",
		"scaleByHostTemplateIDConfig": {"action": "up", "amount": 1.5, "hostTemplateId": "1ht1",  "min": 1, "max": 4, "deleteOption": "mostRecent"}}`)
	request, err = http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler = HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("Amount of type float is invalid")
	}

	jsonStr = []byte(`{"driver":"scaleByHostTemplateID","name":"wh-name",
		"scaleByHostTemplateIDConfig": {"action": "up", "amount": 1, "hostTemplateId": "1ht1",  "min": 1.5, "max": 4, "deleteOption": "mostRecent"}}`)
	request, err = http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler = HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("Min of type float is invalid")
	}

	jsonStr = []byte(`{"driver":"scaleByHostTemplateID","name":"wh-name",
		"scaleByHostTemplateIDConfig": {"action": "up", "amount": 1, "hostTemplateId": "1ht1",  "min": 1, "max": 4.5, "deleteOption": "mostRecent"}}`)
	request, err = http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler = HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("Max of type float is invalid")
	}

	jsonStr = []byte(`{"driver":"scaleByHostTemplateID","name":"wh-name",
		"scaleByHostTemplateIDConfig": {"action": "up", "amount": 1, "hostTemplateId": "1ht1",  "min": 1, "max": 4, "deleteOption": "random"}}`)
	request, err = http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler = HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("Invalid delete option")
	}

	jsonStr = []byte(`{"driver":"scaleByHostTemplateID","name":"wh-name",
		"scaleByHostTemplateIDConfig": {"action": "down", "amount": 1, "hostTemplateId": "1ht1",  "min": 1, "max": 4, "deleteOption": "mostRecent"}}`)
	request, err = http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler = HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("Invalid action")
	}

	jsonStr = []byte(`{"driver":"scaleByHostTemplateID","name":"wh-name",
		"scaleByHostTemplateIDConfig": {"action": "down", "amount": 1, "hostTemplateId": "random",  "min": 1, "max": 4, "deleteOption": "mostRecent"}}`)
	request, err = http.NewRequest("POST", constructURL, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	handler = HandleError(schemas, r.ConstructPayload)
	handler.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("Invalid hostTemplateId")
	}
}

type MockHostWithTemplateIDDriver struct {
	expectedConfig model.ScaleByHostTemplateID
}

func (s *MockHostWithTemplateIDDriver) Execute(conf interface{}, apiClient *client.RancherClient, reqbody interface{}) (int, error) {
	config := &model.ScaleByHostTemplateID{}
	err := mapstructure.Decode(conf, config)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Couldn't unmarshal config: %v", err)
	}

	if config.HostTemplateID != s.expectedConfig.HostTemplateID {
		return 500, fmt.Errorf("HostTemplateID. Expected %v, Actual %v", s.expectedConfig.HostTemplateID, config.HostTemplateID)
	}

	if config.Action != s.expectedConfig.Action {
		return 500, fmt.Errorf("Action. Expected %v, Actual %v", s.expectedConfig.Action, config.Action)
	}

	if config.Amount != s.expectedConfig.Amount {
		return 500, fmt.Errorf("Amount. Expected %v, Actual %v", s.expectedConfig.Amount, config.Amount)
	}

	logrus.Infof("Execute of mock scale host by HostTemplateID driver")
	return 0, nil
}

func (s *MockHostWithTemplateIDDriver) ValidatePayload(conf interface{}, apiClient *client.RancherClient) (int, error) {
	config, ok := conf.(model.ScaleByHostTemplateID)
	if !ok {
		return http.StatusInternalServerError, fmt.Errorf("Can't process config")
	}

	if config.Action != s.expectedConfig.Action {
		return 400, fmt.Errorf("Action. Expected %v, Actual %v", s.expectedConfig.Action, config.Action)
	}

	if config.Amount != s.expectedConfig.Amount {
		return 500, fmt.Errorf("Amount. Expected %v, Actual %v", s.expectedConfig.Amount, config.Amount)
	}

	if config.Min != s.expectedConfig.Min {
		return 500, fmt.Errorf("Min. Expected %v, Actual %v", s.expectedConfig.Min, config.Min)
	}

	if config.Max != s.expectedConfig.Max {
		return 500, fmt.Errorf("Max. Expected %v, Actual %v", s.expectedConfig.Max, config.Max)
	}

	if config.DeleteOption != s.expectedConfig.DeleteOption {
		return 400, fmt.Errorf("Delete option. Expected %v, Actual %v", s.expectedConfig.DeleteOption, config.DeleteOption)
	}

	if config.HostTemplateID != s.expectedConfig.HostTemplateID {
		return 500, fmt.Errorf("HostTemplateID. Expected %v, Actual %v", s.expectedConfig.HostTemplateID, config.HostTemplateID)
	}

	logrus.Infof("Validate payload of mock scale host by HostTemplateID driver")
	return 0, nil
}

func (s *MockHostWithTemplateIDDriver) GetDriverConfigResource() interface{} {
	return model.ScaleByHostTemplateID{}
}

func (s *MockHostWithTemplateIDDriver) CustomizeSchema(schema *v1client.Schema) *v1client.Schema {
	return schema
}

func (s *MockHostWithTemplateIDDriver) ConvertToConfigAndSetOnWebhook(conf interface{}, webhook *model.Webhook) error {
	ss := &drivers.ScaleByHostTemplateIDDriver{}
	return ss.ConvertToConfigAndSetOnWebhook(conf, webhook)
}
