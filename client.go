package abiquo_api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/ernesto-jimenez/httplogger"
	"github.com/go-resty/resty"
	"github.com/technoweenie/multipartstreamer"
)

type AbiquoClient struct {
	client     *resty.Client
	baseClient *http.Client
}

func GetClient(apiurl string, user string, pass string, insecure bool) *AbiquoClient {
	rc := resty.New()

	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	logger := &httpLogger{
		log: log.New(os.Stderr, "AbiquoAPI - ", log.LstdFlags),
	}

	var baseClient *http.Client
	if os.Getenv("ABIQUO_DEBUG") != "" {
		baseClient = &http.Client{
			Transport: httplogger.NewLoggedTransport(baseTransport, logger),
		}
	} else {
		baseClient = &http.Client{
			Transport: baseTransport,
		}
	}

	rc.SetHostURL(apiurl)
	rc.SetBasicAuth(user, pass)
	rc.SetTransport(baseClient.Transport)

	return &AbiquoClient{client: rc}
}

func GetOAuthClient(apiurl string, api_key string, api_secret string, token string, token_secret string, insecure bool) *AbiquoClient {
	rc := resty.New()

	oauth_config := oauth1.NewConfig(api_key, url.QueryEscape(api_secret))
	oauth_token := oauth1.NewToken(token, url.QueryEscape(token_secret))

	baseTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}

	logger := &httpLogger{
		log: log.New(os.Stderr, "AbiquoAPI - ", log.LstdFlags),
	}

	var baseClient *http.Client
	if os.Getenv("ABIQUO_DEBUG") != "" {
		baseClient = &http.Client{
			Transport: httplogger.NewLoggedTransport(baseTransport, logger),
		}
	} else {
		baseClient = &http.Client{
			Transport: baseTransport,
		}
	}

	rc.SetHostURL(apiurl)
	ctx := context.WithValue(oauth1.NoContext, oauth1.HTTPClient, baseClient)
	httpClient := oauth_config.Client(ctx, oauth_token)

	rc.SetTransport(httpClient.Transport)

	return &AbiquoClient{client: rc, baseClient: baseClient}
}

func (c *AbiquoClient) checkResponse(resp *resty.Response, err error) (*resty.Response, error) {
	if err != nil {
		return resp, err
	}

	if resp.StatusCode() > 399 {
		var errCol ErrorCollection
		err = json.Unmarshal(resp.Body(), &errCol)
		if err != nil {
			// Not errorDTO
			abqerror := fmt.Errorf("ERROR %d: %s", resp.StatusCode(), resp.Body())
			return resp, abqerror
		}

		abqerror := errCol.Collection[0]
		err = fmt.Errorf("ERROR %s - %s (HTTP %d)", abqerror.Code, abqerror.Message, resp.StatusCode())
	}
	return resp, err
}

type httpLogger struct {
	log *log.Logger
}

func (l *httpLogger) LogRequest(req *http.Request) {
	l.log.Printf(
		"Request %s %s",
		req.Method,
		req.URL.String(),
	)
	for name, value := range req.Header {
		l.log.Printf("Header '%v': '%v'\n", name, value)
	}
}

func (l *httpLogger) LogResponse(req *http.Request, res *http.Response, err error, duration time.Duration) {
	duration /= time.Millisecond
	if err != nil {
		l.log.Println(err)
	} else {
		l.log.Printf(
			"Response method=%s status=%d durationMs=%d %s",
			req.Method,
			res.StatusCode,
			duration,
			req.URL.String(),
		)
		for name, value := range res.Header {
			l.log.Printf("Header '%v': '%v'\n", name, value)
		}
	}
}

func (c *AbiquoClient) GetEvents(params map[string]string) ([]Event, error) {
	var events []Event
	var eventsCol EventCollection

	events_resp, err := c.client.R().SetHeader("Accept", "application/vnd.abiquo.events+json").
		SetQueryParams(params).
		Get(fmt.Sprintf("%s/events", c.client.HostURL))
	if err != nil {
		return events, err
	}

	err = json.Unmarshal(events_resp.Body(), &eventsCol)
	for {
		for _, e := range eventsCol.Collection {
			events = append(events, e)
		}

		if eventsCol.HasNext() {
			next_link := eventsCol.GetNext()
			events_resp, err = c.client.R().SetHeader("Accept", "application/vnd.abiquo.events+json").
				Get(next_link.Href)
			if err != nil {
				return events, err
			}
			json.Unmarshal(events_resp.Body(), &eventsCol)
		} else {
			break
		}
	}
	return events, nil
}

func (c *AbiquoClient) GetConfigProperties() ([]ConfigProperty, error) {
	var propsCol ConfigPropertyCollection
	var allprops []ConfigProperty

	props_resp, err := c.client.R().SetHeader("Accept", "application/vnd.abiquo.systemproperties+json").
		Get(fmt.Sprintf("%s/config/properties", c.client.HostURL))
	if err != nil {
		return allprops, err
	}

	err = json.Unmarshal(props_resp.Body(), &propsCol)
	if err != nil {
		return allprops, err
	}
	for {
		for _, p := range propsCol.Collection {
			allprops = append(allprops, p)
		}

		if propsCol.HasNext() {
			next_link := propsCol.GetNext()
			props_resp, err = c.client.R().SetHeader("Accept", "application/vnd.abiquo.systemproperties+json").
				Get(next_link.Href)
			if err != nil {
				return allprops, err
			}
			json.Unmarshal(props_resp.Body(), &propsCol)
		} else {
			break
		}
	}
	return allprops, nil
}

func (c *AbiquoClient) GetConfigProperty(name string) (ConfigProperty, error) {
	var prop ConfigProperty
	props, err := c.GetConfigProperties()
	if err != nil {
		return prop, err
	}
	for _, p := range props {
		if p.Name == name {
			return p, nil
		}
	}
	errorMsg := fmt.Sprintf("Property '%s' was not found.", name)
	return prop, errors.New(errorMsg)
}

func (c *AbiquoClient) GetVDCs() ([]VDC, error) {
	var vdcscol VdcCollection
	var allVdcs []VDC

	vdcs_resp, err := c.client.R().SetHeader("Accept", "application/vnd.abiquo.virtualdatacenters+json").
		Get(fmt.Sprintf("%s/cloud/virtualdatacenters", c.client.HostURL))
	if err != nil {
		return allVdcs, err
	}

	err = json.Unmarshal(vdcs_resp.Body(), &vdcscol)
	if err != nil {
		return allVdcs, err
	}
	for {
		for _, v := range vdcscol.Collection {
			allVdcs = append(allVdcs, v)
		}

		if vdcscol.HasNext() {
			next_link := vdcscol.GetNext()
			vdcs_resp, err = c.client.R().SetHeader("Accept", "application/vnd.abiquo.virtualdatacenters+json").
				Get(next_link.Href)
			if err != nil {
				return allVdcs, err
			}
			json.Unmarshal(vdcs_resp.Body(), &vdcscol)
		} else {
			break
		}
	}
	return allVdcs, nil
}

func (c *AbiquoClient) GetVMByUrl(vm_url string) (VirtualMachine, error) {
	var vm VirtualMachine

	vm_raw, err := c.client.R().SetHeader("Accept", "application/vnd.abiquo.virtualmachine+json").
		Get(vm_url)
	if err != nil {
		return vm, err
	}
	if vm_raw.StatusCode() == 404 {
		return vm, errors.New("NOT FOUND")
	}
	json.Unmarshal(vm_raw.Body(), &vm)
	return vm, nil
}

func (c *AbiquoClient) Login() (User, error) {
	var user User
	login_resp, err := c.checkResponse(c.client.R().SetHeader("Accept", "application/vnd.abiquo.user+json").
		Get(fmt.Sprintf("%s/login", c.client.HostURL)))
	if err != nil {
		return user, err
	}
	json.Unmarshal(login_resp.Body(), &user)
	return user, nil
}

func (c *AbiquoClient) upload(uri string, params map[string]string, paramName, path string) (*http.Response, error) {
	// Need the X-Abiquo-Token to authenticate to AM
	login_resp, err := c.checkResponse(c.client.R().SetHeader("Accept", "application/vnd.abiquo.user+json").
		Get(fmt.Sprintf("%s/login", c.client.HostURL)))
	if err != nil {
		return nil, err
	}

	token := ""
	for k, v := range login_resp.RawResponse.Header {
		if k == "X-Abiquo-Token" {
			token = v[0]
		}
	}

	httpReq, err := http.NewRequest("POST", uri, nil)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Token %s", token))
	ms := multipartstreamer.New()
	ms.WriteFields(params)
	ms.WriteFile(paramName, path)
	ms.SetupRequest(httpReq)

	resp, err := c.baseClient.Do(httpReq)
	if resp.StatusCode > 399 {
		var errCol ErrorCollection
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		log.Printf("Response Body : %s", bodyBytes)
		if err != nil {
			return resp, err
		}
		err = json.Unmarshal(bodyBytes, &errCol)
		if err != nil {
			// Not errorDTO
			abqerror := fmt.Errorf("ERROR %d: %s", resp.StatusCode, string(bodyBytes))
			return resp, abqerror
		}

		abqerror := errCol.Collection[0]
		err = fmt.Errorf("ERROR %s - %s (HTTP %d)", abqerror.Code, abqerror.Message, resp.StatusCode)
		return resp, err
	}

	return resp, err
}
