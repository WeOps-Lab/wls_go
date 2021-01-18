package exporter

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/benridley/wls_go/wls"
	"github.com/iancoleman/strcase"
	"github.com/prometheus/client_golang/prometheus"
)

/*
Exporter is the overarching type that contains the configuration and client required to perform
lookups against the Weblogic API
*/
type Exporter struct {
	queryConfig MbeanQuery
	configMap   MBeanConfigMap   // A map of the form <mBeanName, mBeanConfig> for mapping mbeans to labels and metric prefixes
	client      http.Client      // The client used to perform the probing against the Weblogic API
	query       wls.WLSRestQuery // Stores the query required by the exporter to prevent having to recreate it every time
	LogLevel    int              // Which log level to use, 0 = INFO, 1 = DEBUG
}

// MBeanConfig contains the data from config needed to create prometheus metrics from raw mBean data
type MBeanConfig struct {
	LabelName           string          // The label to use for this mBean when converting to Prometheus metrics
	LabelValueAttribute string          // Which attribute of the mBean to use as the label's value
	MetricPrefix        string          // An optional prefix to add to the resultant metrics for organising metrics
	StringFieldInfo     stringFieldInfo // A set that contains mBean attributes which return strings. Used to enumerate all possible labels and provide consistent metrics
}

// MBeanConfigMap is a map of the form <MbeanName, MBeanConfig> so the exporter knows which labels and prefixes to use
// when parsing a Weblogic API response
type MBeanConfigMap map[string]MBeanConfig

/*
WeblogicAPIResponse represents what the Weblogic API returns when queried by the exporter
*/
type WeblogicAPIResponse struct {
	Items           []*WeblogicAPIResponse
	NumericalFields map[string]float64
	StringFields    map[string]string
	ObjectFields    map[string]interface{}
	Children        map[string]*WeblogicAPIResponse
}

/*
StringField represents a field from a Weblogic mBean that returns a string value.
The value set should contain all possible string values that can be returned by the
field so that the exporter can populate all possible labels and provide clean time
series metrics. It's not intended to retrieve arbitrary strings, but rather things
like health states and deployment states that have known potential values.
*/
type StringField struct {
	Name     string   `yaml:"name,omitempty"`
	ValueSet []string `yaml:"value_set,omitempty"`
}

// stringFieldInfo represnts the possible states of a string mBean attribute, converted from StringFields found in config.
// Used a set with labels set to true to indicate their presence.
type stringFieldInfo map[string]map[string]bool

/*
MbeanQuery is the configuration for each desired mbean.
LabelName: This is an optional field that determines the name of the label on the outgoing metric
LabelValueAttribute: Which mbean attribute should be queried for the LabelName value
Fields: Desired attirbutes that return numerical data
StringFields: Desired attributes that return a string. These will be converted to labels with 1 as the current state, 0 as other states.
Children: Child mbeans to also be queried
*/
type MbeanQuery struct {
	LabelName           string                `yaml:"label_name,omitempty"`
	LabelValueAttribute string                `yaml:"label_value_attribute,omitempty"`
	MetricPrefix        string                `yaml:"metric_prefix,omitempty"`
	Fields              []string              `yaml:"fields,omitempty"`
	StringFields        []StringField         `yaml:"string_fields,omitempty"`
	Children            map[string]MbeanQuery `yaml:"children,omitempty"`
}

// Populates a map to easily retrieve each mBean's monitoring config, such as label prefixes and label names
func (cm MBeanConfigMap) createConfigMap(beanName string, q *MbeanQuery) {
	beanConfig := MBeanConfig{
		LabelName:           q.LabelName,
		LabelValueAttribute: q.LabelValueAttribute,
		MetricPrefix:        q.MetricPrefix,
		StringFieldInfo:     make(stringFieldInfo),
	}
	cm[beanName] = beanConfig
	for _, stringField := range q.StringFields {
		beanConfig.StringFieldInfo[stringField.Name] = make(map[string]bool)
		for _, value := range stringField.ValueSet {
			beanConfig.StringFieldInfo[stringField.Name][value] = true
		}
	}
	if q.Children == nil {
		return
	}
	for childName, childConfig := range q.Children {
		cm.createConfigMap(childName, &childConfig)
	}
}

// New creates an exporter from an MBeanQuery and a log level
func New(q MbeanQuery, logLevel int) (Exporter, error) {
	if len(q.Children) == 0 && len(q.Fields) == 0 {
		return Exporter{}, errors.New("Cannot use empty config. No queries specified")
	}
	configMap := MBeanConfigMap{}
	configMap.createConfigMap("serverRuntime", &q)

	query := q.getRESTQuery()

	return Exporter{
		queryConfig: q,
		configMap:   configMap,
		client:      http.Client{Timeout: 10 * time.Second},
		query:       query,
		LogLevel:    logLevel,
	}, nil
}

// UnmarshalYAML implements the yaml.Unmarshaler interface for MbeanQuery
func (q *MbeanQuery) UnmarshalYAML(unmarshal func(interface{}) error) error {
	// Create a type alias to avoid infinite recursion
	type queryYAML MbeanQuery
	qy := (*queryYAML)(q)
	if err := unmarshal(qy); err != nil {
		return err
	}
	if q.LabelName != "" && q.LabelValueAttribute == "" {
		return fmt.Errorf("Cannot parse config at label_name: %s. Must provide label_value_attribute if providing a label_name", q.LabelName)
	} else if q.LabelName == "" && q.LabelValueAttribute != "" {
		return fmt.Errorf("Cannot parse config at label_value_attribute: %s. Must provide label_name if providing a label_value_attribute", q.LabelValueAttribute)
	}
	return nil
}

// GetRESTQueryJSON provides the raw json required to send a request to the WLS API
func (e *Exporter) GetRESTQueryJSON() (json.RawMessage, error) {
	qJSON, err := json.Marshal(&e.query)
	if err != nil {
		return json.RawMessage{}, err
	}
	return qJSON, nil

}

// GetRESTQuery produces the JSON query for each mbean object to be used in the WLS api.
func (q *MbeanQuery) getRESTQuery() wls.WLSRestQuery {
	children := make(map[string]*wls.WLSRestQuery)
	for name, config := range q.Children {
		q := config.getRESTQuery()
		children[name] = &q
	}

	// Set empty array if fields isn't set, otherwise WLS api returns all fields.
	var fields []string
	if len(q.Fields)+len(q.StringFields) != 0 {
		fields = q.Fields
		for _, stringField := range q.StringFields {
			fields = append(fields, stringField.Name)
		}
	} else {
		fields = []string{}
	}

	// Add label value attribute to 'fields' as we need its value implicitly (for labels)
	if q.LabelValueAttribute != "" {
		fields = append(fields, q.LabelValueAttribute)
	}

	return wls.WLSRestQuery{
		Fields:   fields,
		Children: children,
		Links:    []string{},
	}
}

// DoQuery performs a Weblogic query and returns the Prometheus metrics generated from the Weblogic API response
func (e *Exporter) DoQuery(host string, port int, username, password string) ([]prometheus.Gauge, error) {
	queryJSON, err := e.GetRESTQueryJSON()
	if err != nil {
		return nil, err
	}

	basePath := "/management/weblogic/latest/serverRuntime/search"
	path := fmt.Sprintf("http://%s:%d%s", host, port, basePath)

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(queryJSON))
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-Requested-By", "GoWlsClient")
	req.Header.Add("accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(username, password)

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	w := WeblogicAPIResponse{}
	err = json.Unmarshal(body, &w)
	if err != nil {
		return nil, err
	}

	metrics, err := e.CreateMetrics(&w)
	if err != nil {
		return nil, err
	}

	return metrics, err
}

/*
UnmarshalJSON implements the Unmarshaler interface for WeblogicAPIResponse
*/
func (w *WeblogicAPIResponse) UnmarshalJSON(data []byte) error {
	jsonData := map[string]interface{}{}
	json.Unmarshal(data, &jsonData)
	err := w.parseAPIResponse(jsonData)
	if err != nil {
		return err
	}
	return nil
}

/*
Weblogic does not identify 'fields' that are objects from child mBean objects,
so we need to use heuristics to determine if a field is an object or an mBean.
*/
var weblogicObjectFieldNames map[string]bool = map[string]bool{
	"healthState": true,
}

/*
parseAPIResponse unpicks the Weblogic API's JSON response into a proper struct representation
an handles some idiosyncracies of the API.
*/
func (w *WeblogicAPIResponse) parseAPIResponse(data map[string]interface{}) error {
	for key, value := range data {
		switch value := value.(type) {
		case []interface{}:
			if w.Items == nil {
				w.Items = make([]*WeblogicAPIResponse, 0, len(value))
			}
			// Weblogic API will return a list with a single empty object rather than an empty list, so need to check
			// first item to see if a list is empty
			if len(value) == 1 && len(value[0].(map[string]interface{})) == 0 {
				break
			}
			for _, item := range value {
				i, ok := item.(map[string]interface{})
				if !ok {
					return fmt.Errorf("Invalid item type at %v, expected object but got type %T", key, value)
				}
				childItem := WeblogicAPIResponse{}
				err := childItem.parseAPIResponse(i)
				if err != nil {
					return err
				}
				w.Items = append(w.Items, &childItem)
			}
		case float64:
			if w.NumericalFields == nil {
				w.NumericalFields = make(map[string]float64)
			}
			w.NumericalFields[key] = value
		case string:
			if w.StringFields == nil {
				w.StringFields = make(map[string]string)
			}
			w.StringFields[key] = value
		case map[string]interface{}:
			// Check if item is a field or a child mBean
			if _, ok := weblogicObjectFieldNames[key]; ok {
				if w.ObjectFields == nil {
					w.ObjectFields = make(map[string]interface{})
				}
				w.ObjectFields[key] = value
			} else {
				if w.Children == nil {
					w.Children = make(map[string]*WeblogicAPIResponse)
				}
				childBean := WeblogicAPIResponse{}
				err := childBean.parseAPIResponse(value)
				if err != nil {
					return err
				}
				w.Children[key] = &childBean
			}
		}
	}
	return nil
}

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

/*
CreateMetrics uses an exporter to create metrics from a Weblogic API reseponse
*/
func (e *Exporter) CreateMetrics(resp *WeblogicAPIResponse) (metrics []prometheus.Gauge, err error) {
	// Start at serverRuntime, which is the root node of Weblogic's runtime mBean tree.
	serverMetrics, err := e.createMBeanMetrics("serverRuntime", resp, nil)
	if err != nil {
		return nil, err
	}
	return serverMetrics, nil
}

/*
CreateMBeanMetrics creates a series of Prometheus metrics from an mBean name, a set of labels, and an API response that contains
the metrics for that mBean. It also recursively creates child metrics.
*/
func (e *Exporter) createMBeanMetrics(beanName string, resp *WeblogicAPIResponse, labels prometheus.Labels) (metrics []prometheus.Gauge, err error) {
	metricConfig, ok := e.configMap[beanName]
	if !ok {
		return nil, fmt.Errorf("Unable to find monitoring config for mBean %s", beanName)
	}
	beanLabels := make(prometheus.Labels)
	mainLabelValue, ok := resp.StringFields[metricConfig.LabelValueAttribute]
	if ok {
		beanLabels[metricConfig.LabelName] = mainLabelValue
	}

	// Add extra labels from parameter
	copyLabels(beanLabels, labels)

	metrics = make([]prometheus.Gauge, 0, len(resp.NumericalFields))

	for fieldName, fieldValue := range resp.NumericalFields {
		gauge := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        metricConfig.MetricPrefix + strcase.ToSnake(fieldName),
			ConstLabels: beanLabels,
		})
		gauge.Set(fieldValue)
		metrics = append(metrics, gauge)
	}

	// Create string label metrics. These are similar to systemd metrics in the Node Exporter where all states are enumerated with different labels
	for fieldName, potentialValues := range metricConfig.StringFieldInfo {
		if responseValue, ok := resp.StringFields[fieldName]; ok {
			// Create metrics that represent all the possible string responses set as labels
			fieldLabels := make(prometheus.Labels)
			copyLabels(fieldLabels, beanLabels)
			labelName := strcase.ToSnake(fieldName)
			for potentialValue := range potentialValues {
				fieldLabels[labelName] = potentialValue
				gauge := prometheus.NewGauge(prometheus.GaugeOpts{
					Name:        metricConfig.MetricPrefix + strcase.ToSnake(fieldName),
					ConstLabels: fieldLabels,
				})
				if potentialValue == responseValue {
					gauge.Set(1)
				} else {
					gauge.Set(0)
				}
				metrics = append(metrics, gauge)
			}
		}
	}

	// Handle healthstate metrics, which have standard string outputs that can be
	// easily converted to labels
	if hs, ok := resp.ObjectFields["healthState"]; ok {
		hsw := hs.(map[string]interface{})
		hsString := hsw["state"].(string)

		states := []string{"ok", "overloaded", "warn", "critical", "failed"}
		for _, s := range states {
			stateLabels := prometheus.Labels{}
			copyLabels(stateLabels, beanLabels)
			stateLabels["state"] = s
			gauge := prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        metricConfig.MetricPrefix + "health_state",
				ConstLabels: stateLabels,
			})
			if s == hsString {
				gauge.Set(1)
			} else {
				gauge.Set(0)
			}
			metrics = append(metrics, gauge)
		}
	}

	// Recursively create child metrics
	for _, item := range resp.Items {
		itemMetrics, err := e.createMBeanMetrics(beanName, item, beanLabels)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, itemMetrics...)
	}

	for childName, child := range resp.Children {
		childMetrics, err := e.createMBeanMetrics(childName, child, beanLabels)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, childMetrics...)
	}
	return metrics, nil
}

func copyLabels(new, old prometheus.Labels) {
	for k, v := range old {
		new[k] = v
	}
}
