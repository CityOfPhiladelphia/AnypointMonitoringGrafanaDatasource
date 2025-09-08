package plugin

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/instancemgmt"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"github.com/grafana/grafana-plugin-sdk-go/data"
	"github.com/weastie/anypoint-monitoring/pkg/models"
)

// Make sure Datasource implements required interfaces. This is important to do
// since otherwise we will only get a not implemented error response from plugin in
// runtime. In this example datasource instance implements backend.QueryDataHandler,
// backend.CheckHealthHandler interfaces. Plugin should not implement all these
// interfaces - only those which are required for a particular task.
var (
	_ backend.QueryDataHandler      = (*Datasource)(nil)
	_ backend.CheckHealthHandler    = (*Datasource)(nil)
	_ instancemgmt.InstanceDisposer = (*Datasource)(nil)
)

// NewDatasource creates a new datasource instance.
func NewDatasource(ctx context.Context, _ backend.DataSourceInstanceSettings) (instancemgmt.Instance, error) {
	// Create our HTTP client
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // This line disables SSL verification
	}
	// Create our query template
	queryStringTemplateRaw := `SELECT mean("{{ .metric_name }}") FROM "{{ .metric_table }}" WHERE ("org_id" = '{{ .org_id }}' AND "env_id" = '{{ .env_id }}' AND "cluster_id" = '{{ .cluster_id }}' AND "app_id" = '{{ .app_id }}') AND time >= {{ .start_time }}ms and time <= {{ .end_time }}ms GROUP BY time({{ .time_step }}), "worker_id" fill(null) tz('America/New_York')`

	// Create template
	queryStringTemplate, err := template.New("sql").Parse(queryStringTemplateRaw)
	if err != nil {
		panic(err)
	}

	return &Datasource{
		apiKey:              "",
		apiKeyTime:          time.Time{},
		httpClient:          &http.Client{Transport: tr},
		queryStringTemplate: queryStringTemplate,
		logger:              log.DefaultLogger.FromContext(ctx),
	}, nil
}

// Datasource is an example datasource which can respond to data queries, reports
// its health and has streaming skills.
type Datasource struct {
	apiKey              string
	apiKeyTime          time.Time
	httpClient          *http.Client
	queryStringTemplate *template.Template
	logger              log.Logger
}

// Dispose here tells plugin SDK that plugin wants to clean up resources when a new instance
// created. As soon as datasource settings change detected by SDK old datasource instance will
// be disposed and a new one will be created using NewSampleDatasource factory function.
func (d *Datasource) Dispose() {
	// Clean up datasource instance resources.
}

// QueryData handles multiple queries and returns multiple responses.
// req contains the queries []DataQuery (where each query contains RefID as a unique identifier).
// The QueryDataResponse contains a map of RefID to the response for each query, and each response
// contains Frames ([]*Frame).
func (d *Datasource) QueryData(ctx context.Context, req *backend.QueryDataRequest) (*backend.QueryDataResponse, error) {
	// create response struct
	response := backend.NewQueryDataResponse()

	// loop over queries and execute them individually.
	for _, q := range req.Queries {
		res := d.query(ctx, req.PluginContext, q)

		// save the response in a hashmap
		// based on with RefID as identifier
		response.Responses[q.RefID] = res
	}

	return response, nil
}

// Data structure for token
type apTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Data structure of the JSON response (AP short for Anypoint)
type apResponse struct {
	Results []apResult `json:"results"`
}

type apResult struct {
	Series []apSeries `json:"series"`
}

type apSeries struct {
	Name    string            `json:"name"`
	Columns []string          `json:"columns"`
	Tags    map[string]string `json:"tags"`
	Values  [][]interface{}   `json:"values"`
}

// Data structure of the inputs
type queryModel struct {
	OrgId       string `json:"orgId"`
	EnvId       string `json:"envId"`
	ClusterId   string `json:"clusterId"`
	AppId       string `json:"appId"`
	MetricName  string `json:"metricName"`
	MetricTable string `json:"metricTable"`
	TimeStep    string `json:"timeStep"`
}

func (d *Datasource) generateAPIKey(clientId string, clientSecret string) bool {
	d.logger.Info("Generating new API key")
	tokenURL := "https://anypoint.mulesoft.com/accounts/api/v2/oauth2/token"

	form := url.Values{}
	form.Add("client_id", clientId)
	form.Add("client_secret", clientSecret)
	form.Add("grant_type", "client_credentials")

	// d.logger.Info("Form data", "req", form.Encode())

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		d.logger.Error("Error forming http request", "err", err)
		return false
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		d.logger.Error("Error sending API token request", "err", err)
		return false
	}
	defer resp.Body.Close()

	var apTokenResp apTokenResponse
	body, _ := io.ReadAll(resp.Body)
	// d.logger.Info("Token response", "res", string(body))
	if err := json.Unmarshal(body, &apTokenResp); err != nil {
		d.logger.Error("Error parsing API token response", "err", err)
		return false
	}

	d.apiKey = apTokenResp.AccessToken
	d.apiKeyTime = time.Now()
	return true
}

// Simple wrapper function that returns the current api key if less than 30 minutes
// old, otherwise generates a new one
func (d *Datasource) getAPIKey(clientId string, clientSecret string) string {
	// If it has been less than 30 minutes, just return the key
	if time.Since(d.apiKeyTime).Minutes() < 30 {
		d.logger.Info("Using cached API key")
		return d.apiKey
	}
	// Need to get a new API key
	d.generateAPIKey(clientId, clientSecret)
	return d.apiKey
}

// Internal function getting API key
func (d *Datasource) query(ctx context.Context, pCtx backend.PluginContext, query backend.DataQuery) backend.DataResponse {
	var response backend.DataResponse

	// Unmarshal the JSON into our queryModel.
	var qm queryModel

	err := json.Unmarshal(query.JSON, &qm)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusBadRequest, fmt.Sprintf("json unmarshal: %v", err.Error()))
	}

	// Load plugin settings
	config, err := models.LoadPluginSettings(*pCtx.DataSourceInstanceSettings)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusInternal, "Could not load plugin settings")
	}
	// create data frame response.
	// For an overview on data frames and how grafana handles them:
	// https://grafana.com/developers/plugin-tools/introduction/data-frames

	// Prepare request to anypoint monitoring
	baseURL := "https://anypoint.mulesoft.com/monitoring/api/visualizer/api/datasources/proxy/148362/query"
	queryParamTemplateValues := map[string]interface{}{
		"org_id":       qm.OrgId,
		"env_id":       qm.EnvId,
		"cluster_id":   qm.ClusterId,
		"app_id":       qm.AppId,
		"metric_name":  qm.MetricName,
		"metric_table": qm.MetricTable,
		"time_step":    qm.TimeStep,
		"start_time":   query.TimeRange.From.UnixNano() / int64(time.Millisecond),
		"end_time":     query.TimeRange.To.UnixNano() / int64(time.Millisecond),
	}

	// Construct our payload
	queryParams := url.Values{}
	queryParams.Set("db", "\"dias_mt_19_prod\"")
	queryParams.Set("epoch", "ms")

	// buf for the template execution output
	var buf bytes.Buffer
	err = d.queryStringTemplate.Execute(&buf, queryParamTemplateValues)
	if err != nil {
		return backend.ErrDataResponse(backend.StatusValidationFailed, "Malformed Inputs")
	}

	// Finally we have our parsed queryString (I miss Python)
	queryString := buf.String()

	queryParams.Set("q", queryString)

	// Build the final URL
	u, err := url.Parse(baseURL)
	if err != nil {
		panic(err)
	}

	// Go encodes spaces as + but I want %20
	u.RawQuery = strings.ReplaceAll(queryParams.Encode(), "+", "%20")

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		d.logger.Error("Could not form HTTP request", "err", err)
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.getAPIKey(config.Secrets.ClientId, config.Secrets.ClientSecret))

	// Send the request
	resp, err := d.httpClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Parse the response
	var apResp apResponse
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &apResp); err != nil {
		return backend.ErrDataResponse(backend.StatusUnknown, "Unknown error when parsing Anypoint output")
	}

	// Now, get each data point
	for _, result := range apResp.Results {
		for _, series := range result.Series {
			labelsMap := data.Labels{
				"metric":     series.Name + "." + qm.MetricName,
				"worker":     series.Tags["worker_id"],
				"org_id":     qm.OrgId,
				"env_id":     qm.EnvId,
				"cluster_id": qm.ClusterId,
				"app_id":     qm.AppId,
			}

			frame := data.NewFrame(labelsMap.String())
			for columnId, columnName := range series.Columns {
				var dataField []float64
				var dataTimes []time.Time
				for _, row := range series.Values {
					// Ignore empty rows if they happen
					if len(row) < 2 {
						continue
					}
					// Time is special, we have to parse everything into a time
					if columnName == "time" {
						tsFloat, ok := row[0].(float64)
						if !ok {
							// Append a 0 like value if it can't be parsed
							dataTimes = append(dataTimes, time.Time{})
						} else {
							// Convert ms to time.Time
							ts := time.Unix(0, int64(tsFloat)*int64(time.Millisecond))
							dataTimes = append(dataTimes, ts)
						}
					} else {
						// Everything else is just a float
						// Assert the value is a float
						dataValue, ok := row[columnId].(float64)
						if !ok {
							dataField = append(dataField, math.NaN())
						} else {
							dataField = append(dataField, dataValue)
						}
					}
				}
				if columnName == "time" {
					frame.Fields = append(frame.Fields,
						data.NewField(columnName, labelsMap, dataTimes),
					)
				} else {
					frame.Fields = append(frame.Fields,
						data.NewField(columnName, labelsMap, dataField),
					)
				}
			}
			response.Frames = append(response.Frames, frame)
		}
	}

	// logger.Warn("URL: " + u.String())
	// logger.Warn("body: " + string(body))

	// add the frames to the response.

	return response
}

// CheckHealth handles health checks sent from Grafana to the plugin.
// The main use case for these health checks is the test button on the
// datasource configuration page which allows users to verify that
// a datasource is working as expected.
func (d *Datasource) CheckHealth(_ context.Context, req *backend.CheckHealthRequest) (*backend.CheckHealthResult, error) {
	res := &backend.CheckHealthResult{}
	config, err := models.LoadPluginSettings(*req.PluginContext.DataSourceInstanceSettings)
	if err != nil {
		res.Status = backend.HealthStatusError
		res.Message = "Unable to load settings"
		return res, nil
	}

	if d.generateAPIKey(config.Secrets.ClientId, config.Secrets.ClientSecret) {
		return &backend.CheckHealthResult{
			Status:  backend.HealthStatusOk,
			Message: "Data source is working",
		}, nil
	} else {
		res.Status = backend.HealthStatusError
		res.Message = "Failed retrieving API key"
		return res, nil
	}
}
