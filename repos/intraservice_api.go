package repos

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"git.countmax.ru/countmax/commonapi/domain"

	"go.uber.org/zap"
)

const (
	maxIdleConns int = 10
)

// ISAPI repository for servicedesk entities
type ISAPI struct {
	url     string
	user    string
	pass    string
	handler *http.Client
	log     *zap.SugaredLogger
}

// NewISAPI builder for ISAPI
func NewISAPI(url, user, pass string, timeout time.Duration, logger *zap.SugaredLogger) *ISAPI {
	logger = logger.With(zap.String("sd_url", url))
	api := &ISAPI{
		url:  url,
		user: user,
		pass: pass,
		log:  logger,
	}
	// nolint:gosec
	// accept google invalid *.watcom.ru cert error
	cfg := &tls.Config{
		InsecureSkipVerify: true,
	}
	// set timeouts and border for opened connections
	api.handler = &http.Client{Timeout: timeout, Transport: &http.Transport{
		MaxIdleConns:    maxIdleConns,
		IdleConnTimeout: timeout,
		TLSClientConfig: cfg,
	}}
	return api
}

func (api *ISAPI) FindAllBySN(serNomer string, offset, limit int64) (domain.ControllerTasks, int64, error) {
	uri := fmt.Sprintf("%s/api/task?search=%s&pagesize=%d&page=1", api.url, serNomer, limit)
	log := api.log.With(zap.String("target", uri))
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Errorf("make NewRequest error, %v", err)
		return nil, 0, err
	}
	request.SetBasicAuth(api.user, api.pass)
	response, err := api.handler.Do(request)
	if err != nil {
		log.Errorf("get response error, %v", err)
		return nil, 0, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		txt, er := ioutil.ReadAll(response.Body)
		if er != nil {
			log.Errorf("ioutil.ReadAll error, %s", er)
		}
		log.Warnf("task search wrong response, status=%s, body=%s", response.Status, string(txt))
		return nil, 0, fmt.Errorf("task search wrong response, status=%s, body=%s", response.Status, string(txt))
	}
	rawTasks := make(map[string]interface{})
	err = json.NewDecoder(response.Body).Decode(&rawTasks)
	if err != nil {
		log.Errorf("make json.NewDecoder.Decode error, %v", err)
		return nil, 0, err
	}
	pagi := rawTasks["Paginator"].(map[string]interface{})
	count := pagi["Count"].(float64)

	tasks, ok := rawTasks["Tasks"].([]interface{})
	if !ok {
		log.Error("make slice of Tasks error")
		return nil, 0, fmt.Errorf("make slice of Tasks error")
	}
	result := make(domain.ControllerTasks, 0, len(tasks))
	for _, v := range tasks {
		json, err := json.Marshal(v)
		if err != nil {
			continue
		}
		item := domain.ControllerTask{
			Task: string(json),
		}
		result = append(result, item)
	}
	return result, int64(count), nil
}

func (api *ISAPI) TaskAddComment(taskID string, comment string) error {
	uri := fmt.Sprintf("%s/api/task/%s", api.url, taskID)
	log := api.log.With(zap.String("target", uri))

	params := map[string]string{
		"Comment": comment,
	}

	bodyrequest, err := json.Marshal(params)
	if err != nil {
		log.Errorf("make body request error, %v", err)
		return err
	}
	// Set reader from body request
	r := bytes.NewReader(bodyrequest)
	// make request
	request, err := http.NewRequest("PUT", uri, r)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	// add authorization headers
	request.SetBasicAuth(api.user, api.pass)
	response, err := api.handler.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		txt, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Errorf("ioutil.ReadAll error, %v", err)
		}
		err = fmt.Errorf("http response status=%s, code=%d, response=%s", response.Status, response.StatusCode, string(txt))
		return err
	}
	return nil
}

func (api *ISAPI) TaskSetStatus(taskID string, status domain.TaskStatus) error {
	uri := fmt.Sprintf("%s/api/task/%s", api.url, taskID)
	log := api.log.With(zap.String("target", uri))

	bodyParams := map[string]string{}
	if status.Comment != "" {
		bodyParams["Comment"] = status.Comment
		bodyParams["IsPrivateComment"] = strconv.FormatBool(status.IsPrivateComment)
	}
	if status.StatusID != 0 {
		bodyParams["StatusId"] = strconv.Itoa(status.StatusID)
	}
	if status.ResultFieldName != "" {
		bodyParams[status.ResultFieldName] = status.ResultFieldValue
	}

	bodyrequest, err := json.Marshal(bodyParams)
	if err != nil {
		log.Errorf("make body request error, %v", err)
		return err
	}
	// Set reader from body request
	r := bytes.NewReader(bodyrequest)
	// make request
	request, err := http.NewRequest("PUT", uri, r)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	// add authorization headers
	request.SetBasicAuth(api.user, api.pass)
	response, err := api.handler.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		txt, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Errorf("ioutil.ReadAll error, %v", err)
		}
		return fmt.Errorf("http response status=%s, code=%d, response=%s", response.Status, response.StatusCode, string(txt))
	}
	return nil
}

func (api *ISAPI) Health() error {
	uri := fmt.Sprintf("%s/api/tasktype?fields=Id", api.url)
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}
	request.SetBasicAuth(api.user, api.pass)
	response, err := api.handler.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("status code %d", response.StatusCode)
	}
	return nil
}
