package exporter

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

var configTestCases = []struct {
	configPath string
	queries    MbeanQuery
	expectErr  bool
}{
	{
		configPath: "testdata/basic_config_good.yml",
		queries: MbeanQuery{
			LabelName:           "server",
			LabelValueAttribute: "name",
			Fields:              []string{"healthState"},
			Children: map[string]MbeanQuery{
				"JVMRuntime": {
					LabelName:           "",
					LabelValueAttribute: "",
					Fields:              []string{"heapFreeCurrent"},
					Children:            nil,
				},
			},
		},
		expectErr: false,
	},
	{
		configPath: "testdata/empty_config_bad.yml",
		queries:    MbeanQuery{},
		expectErr:  true,
	},
	{
		configPath: "testdata/advanced_config_good.yml",
		queries: MbeanQuery{
			LabelName:           "server",
			LabelValueAttribute: "name",
			Fields:              nil,
			Children: map[string]MbeanQuery{
				"JVMRuntime": {
					LabelName:           "",
					LabelValueAttribute: "",
					Fields:              []string{"heapFreeCurrent", "heapFreePercent", "heapSizeCurrent"},
					Children:            nil,
					MetricPrefix:        "wls_jvm_",
				},
				"applicationRuntimes": {
					LabelName:           "application_runtime",
					LabelValueAttribute: "name",
					Children: map[string]MbeanQuery{
						"componentRuntimes": {
							LabelName:           "component_runtime",
							LabelValueAttribute: "name",
							Fields:              []string{"deploymentState", "sessionsOpenedTotalCount"},
							MetricPrefix:        "wls_webapp_",
							Children: map[string]MbeanQuery{
								"servlets": {
									LabelName:           "servlet",
									LabelValueAttribute: "servletName",
									Fields:              []string{"invocationTotalCount", "executionTimeAverage", "executionTimeHigh", "executionTimeTotal"},
									MetricPrefix:        "wls_servlet_",
									Children:            nil,
								},
							},
						},
					},
				},
				"JDBCServiceRuntime": {
					LabelName:           "jdbc_service",
					LabelValueAttribute: "name",
					Children: map[string]MbeanQuery{
						"JDBCDataSourceRuntimeMBeans": {
							LabelName:           "datasource",
							LabelValueAttribute: "name",
							Fields:              []string{"connectionsTotalCount", "deploymentState", "currCapacity"},
							MetricPrefix:        "wls_datasource_",
						},
					},
				},
				"threadPoolRuntime": {
					LabelName:           "threadpool",
					LabelValueAttribute: "name",
					Fields:              []string{"stuckThreadCount"},
					MetricPrefix:        "wls_threadpool_",
				},
			},
		},
		expectErr: false,
	},
}

var responseTestCases = []struct {
	queries        MbeanQuery
	apiResponse    string
	expectError    bool
	parsedResponse WeblogicAPIResponse
}{
	{
		queries:     configTestCases[0].queries,
		apiResponse: "{\"healthState\":{\"state\":\"ok\",\"subsystemName\":null,\"partitionName\":null,\"symptoms\":[]},\"name\":\"admin-server\",\"JVMRuntime\":{\"heapFreeCurrent\":71934392}}",
		expectError: false,
		parsedResponse: WeblogicAPIResponse{
			ObjectFields: map[string]interface{}{
				"healthState": map[string]interface{}{
					"state":         "ok",
					"subsystemName": nil,
					"partitionName": nil,
					"symptoms":      []interface{}{},
				},
			},
			StringFields: map[string]string{"name": "admin-server"},
			Children: map[string]*WeblogicAPIResponse{
				"JVMRuntime": {
					NumericalFields: map[string]float64{"heapFreeCurrent": 71934392},
				},
			},
		},
	},
	{
		queries:     configTestCases[2].queries,
		apiResponse: "{\"name\":\"admin-server\",\"threadPoolRuntime\":{\"stuckThreadCount\":0},\"JDBCServiceRuntime\":{\"JDBCDataSourceRuntimeMBeans\":{\"items\":[]}},\"JVMRuntime\":{\"heapSizeCurrent\":201068544,\"heapFreeCurrent\":99662296,\"heapFreePercent\":56},\"applicationRuntimes\":{\"items\":[{\"componentRuntimes\":{\"items\":[{\"deploymentState\":2,\"sessionsOpenedTotalCount\":19,\"name\":\"admin-server_\\/management\",\"servlets\":{\"items\":[{\"executionTimeHigh\":0,\"servletName\":\"JspServlet\",\"invocationTotalCount\":272,\"executionTimeTotal\":2465,\"executionTimeAverage\":180},{\"executionTimeHigh\":223,\"servletName\":\"FileServlet\",\"invocationTotalCount\":0,\"executionTimeTotal\":0,\"executionTimeAverage\":0}]}}]}},{\"componentRuntimes\":{\"items\":[{\"deploymentState\":2,\"name\":\"jms-internal-notran-adp\"}]}},{\"componentRuntimes\":{\"items\":[{\"deploymentState\":2,\"sessionsOpenedTotalCount\":0,\"name\":\"admin-server_\\/bea_wls_internal\",\"servlets\":{\"items\":[{\"executionTimeHigh\":0,\"servletName\":\"HTTPClntSend\",\"invocationTotalCount\":0,\"executionTimeTotal\":0,\"executionTimeAverage\":0},{\"executionTimeHigh\":0,\"servletName\":\"HTTPClntClose\",\"invocationTotalCount\":0,\"executionTimeTotal\":0,\"executionTimeAverage\":0}]}}]}},{\"componentRuntimes\":{\"items\":[{\"deploymentState\":2,\"sessionsOpenedTotalCount\":1,\"name\":\"admin-server_\\/console\",\"servlets\":{\"items\":[{\"executionTimeHigh\":0,\"servletName\":\"JspServlet\",\"invocationTotalCount\":0,\"executionTimeTotal\":0,\"executionTimeAverage\":0},{\"executionTimeHigh\":57,\"servletName\":\"\\/login\\/LoginForm.jsp\",\"invocationTotalCount\":1,\"executionTimeTotal\":57,\"executionTimeAverage\":57}]}},{\"deploymentState\":2,\"sessionsOpenedTotalCount\":0,\"name\":\"admin-server_\\/consolehelp\",\"servlets\":{\"items\":[{\"executionTimeHigh\":0,\"servletName\":\"JspServlet\",\"invocationTotalCount\":0,\"executionTimeTotal\":0,\"executionTimeAverage\":0},{\"executionTimeHigh\":0,\"servletName\":\"portalDependencyServlet\",\"invocationTotalCount\":0,\"executionTimeTotal\":0,\"executionTimeAverage\":0},{\"executionTimeHigh\":0,\"servletName\":\"FileDefault\",\"invocationTotalCount\":0,\"executionTimeTotal\":0,\"executionTimeAverage\":0}]}}]}},{\"componentRuntimes\":{\"items\":[{\"deploymentState\":2,\"name\":\"mejb\"}]}},{\"componentRuntimes\":{\"items\":[{\"deploymentState\":2,\"name\":\"jms-internal-xa-adp\"}]}},{\"componentRuntimes\":{\"items\":[{\"deploymentState\":2,\"sessionsOpenedTotalCount\":0,\"name\":\"admin-server_\\/weblogic\",\"servlets\":{\"items\":[{\"executionTimeHigh\":0,\"servletName\":\"JspServlet\",\"invocationTotalCount\":0,\"executionTimeTotal\":0,\"executionTimeAverage\":0},{\"executionTimeHigh\":32,\"servletName\":\"ready\",\"invocationTotalCount\":1,\"executionTimeTotal\":32,\"executionTimeAverage\":32}]}}]}}]}}",
		expectError: false,
		parsedResponse: WeblogicAPIResponse{
			StringFields: map[string]string{"name": "admin-server"},
			Children: map[string]*WeblogicAPIResponse{
				"JVMRuntime": {
					NumericalFields: map[string]float64{"heapFreeCurrent": 99662296, "heapSizeCurrent": 201068544, "heapFreePercent": 56},
				},
				"threadPoolRuntime": {
					NumericalFields: map[string]float64{"stuckThreadCount": 0},
				},
				"JDBCServiceRuntime": {
					Children: map[string]*WeblogicAPIResponse{
						"JDBCDataSourceRuntimeMBeans": {
							Items: []*WeblogicAPIResponse{},
						},
					},
				},
				"applicationRuntimes": {
					Items: []*WeblogicAPIResponse{
						{
							Children: map[string]*WeblogicAPIResponse{
								"componentRuntimes": {
									Items: []*WeblogicAPIResponse{
										{
											NumericalFields: map[string]float64{"deploymentState": 2, "sessionsOpenedTotalCount": 19},
											StringFields:    map[string]string{"name": "admin-server_/management"},
											Children: map[string]*WeblogicAPIResponse{
												"servlets": {
													Items: []*WeblogicAPIResponse{
														{
															StringFields:    map[string]string{"servletName": "JspServlet"},
															NumericalFields: map[string]float64{"executionTimeHigh": 0, "invocationTotalCount": 272, "executionTimeTotal": 2465, "executionTimeAverage": 180},
														},
														{
															StringFields:    map[string]string{"servletName": "FileServlet"},
															NumericalFields: map[string]float64{"executionTimeHigh": 223, "invocationTotalCount": 0, "executionTimeTotal": 0, "executionTimeAverage": 0},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Children: map[string]*WeblogicAPIResponse{
								"componentRuntimes": {
									Items: []*WeblogicAPIResponse{
										{
											NumericalFields: map[string]float64{"deploymentState": 2},
											StringFields:    map[string]string{"name": "jms-internal-notran-adp"},
										},
									},
								},
							},
						},
						{
							Children: map[string]*WeblogicAPIResponse{
								"componentRuntimes": {
									Items: []*WeblogicAPIResponse{
										{
											NumericalFields: map[string]float64{"deploymentState": 2, "sessionsOpenedTotalCount": 0},
											StringFields:    map[string]string{"name": "admin-server_/bea_wls_internal"},
											Children: map[string]*WeblogicAPIResponse{
												"servlets": {
													Items: []*WeblogicAPIResponse{
														{
															StringFields:    map[string]string{"servletName": "HTTPClntSend"},
															NumericalFields: map[string]float64{"executionTimeHigh": 0, "invocationTotalCount": 0, "executionTimeTotal": 0, "executionTimeAverage": 0},
														},
														{
															StringFields:    map[string]string{"servletName": "HTTPClntClose"},
															NumericalFields: map[string]float64{"executionTimeHigh": 0, "invocationTotalCount": 0, "executionTimeTotal": 0, "executionTimeAverage": 0},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Children: map[string]*WeblogicAPIResponse{
								"componentRuntimes": {
									Items: []*WeblogicAPIResponse{
										{
											NumericalFields: map[string]float64{"deploymentState": 2, "sessionsOpenedTotalCount": 1},
											StringFields:    map[string]string{"name": "admin-server_/console"},
											Children: map[string]*WeblogicAPIResponse{
												"servlets": {
													Items: []*WeblogicAPIResponse{
														{
															StringFields:    map[string]string{"servletName": "JspServlet"},
															NumericalFields: map[string]float64{"executionTimeHigh": 0, "invocationTotalCount": 0, "executionTimeTotal": 0, "executionTimeAverage": 0},
														},
														{
															StringFields:    map[string]string{"servletName": "/login/LoginForm.jsp"},
															NumericalFields: map[string]float64{"executionTimeHigh": 57, "invocationTotalCount": 1, "executionTimeTotal": 57, "executionTimeAverage": 57},
														},
													},
												},
											},
										},
										{
											NumericalFields: map[string]float64{"deploymentState": 2, "sessionsOpenedTotalCount": 0},
											StringFields:    map[string]string{"name": "admin-server_/consolehelp"},
											Children: map[string]*WeblogicAPIResponse{
												"servlets": {
													Items: []*WeblogicAPIResponse{
														{
															StringFields:    map[string]string{"servletName": "JspServlet"},
															NumericalFields: map[string]float64{"executionTimeHigh": 0, "invocationTotalCount": 0, "executionTimeTotal": 0, "executionTimeAverage": 0},
														},
														{
															StringFields:    map[string]string{"servletName": "portalDependencyServlet"},
															NumericalFields: map[string]float64{"executionTimeHigh": 0, "invocationTotalCount": 0, "executionTimeTotal": 0, "executionTimeAverage": 0},
														},
														{
															StringFields:    map[string]string{"servletName": "FileDefault"},
															NumericalFields: map[string]float64{"executionTimeHigh": 0, "invocationTotalCount": 0, "executionTimeTotal": 0, "executionTimeAverage": 0},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						{
							Children: map[string]*WeblogicAPIResponse{
								"componentRuntimes": {
									Items: []*WeblogicAPIResponse{
										{
											NumericalFields: map[string]float64{"deploymentState": 2},
											StringFields:    map[string]string{"name": "mejb"},
										},
									},
								},
							},
						},
						{
							Children: map[string]*WeblogicAPIResponse{
								"componentRuntimes": {
									Items: []*WeblogicAPIResponse{
										{
											NumericalFields: map[string]float64{"deploymentState": 2},
											StringFields:    map[string]string{"name": "jms-internal-xa-adp"},
										},
									},
								},
							},
						},
						{
							Children: map[string]*WeblogicAPIResponse{
								"componentRuntimes": {
									Items: []*WeblogicAPIResponse{
										{
											NumericalFields: map[string]float64{"deploymentState": 2, "sessionsOpenedTotalCount": 0},
											StringFields:    map[string]string{"name": "admin-server_/weblogic"},
											Children: map[string]*WeblogicAPIResponse{
												"servlets": {
													Items: []*WeblogicAPIResponse{
														{
															StringFields:    map[string]string{"servletName": "JspServlet"},
															NumericalFields: map[string]float64{"executionTimeHigh": 0, "invocationTotalCount": 0, "executionTimeTotal": 0, "executionTimeAverage": 0},
														},
														{
															StringFields:    map[string]string{"servletName": "ready"},
															NumericalFields: map[string]float64{"executionTimeHigh": 32, "invocationTotalCount": 1, "executionTimeTotal": 32, "executionTimeAverage": 32},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	},
}

type metricTestSpec struct {
	name   string
	labels map[string]string
	value  float64
}

var metricTestCases = []struct {
	queries        MbeanQuery
	parsedResponse WeblogicAPIResponse
	metrics        []metricTestSpec
	expectErr      bool
}{
	{
		queries:        configTestCases[0].queries,
		parsedResponse: responseTestCases[0].parsedResponse,
		metrics: []metricTestSpec{
			{
				name:   "health_state",
				labels: map[string]string{"server": "admin-server", "state": "ok"},
				value:  float64(1),
			},
			{
				name:   "health_state",
				labels: map[string]string{"server": "admin-server", "state": "overloaded"},
				value:  float64(0),
			},
			{
				name:   "health_state",
				labels: map[string]string{"server": "admin-server", "state": "warn"},
				value:  float64(0),
			},
			{
				name:   "health_state",
				labels: map[string]string{"server": "admin-server", "state": "critical"},
				value:  float64(0),
			},
			{
				name:   "health_state",
				labels: map[string]string{"server": "admin-server", "state": "failed"},
				value:  float64(0),
			},
			{
				name:   "heap_free_current",
				labels: map[string]string{"server": "admin-server"},
				value:  float64(71934392),
			},
		},
		expectErr: false,
	},
}

func TestUnmarshalResponse(t *testing.T) {
	for _, testCase := range responseTestCases {
		resp := WeblogicAPIResponse{}
		err := json.Unmarshal([]byte(testCase.apiResponse), &resp)
		if err != nil && !testCase.expectError {
			t.Error(err)
		}
		if !reflect.DeepEqual(resp, testCase.parsedResponse) {
			t.Errorf("Want %s\nGot %s\n", prettyPrint(testCase.parsedResponse), prettyPrint(resp))
		}
	}
}

func BenchmarkParseResponse(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, testCase := range responseTestCases {
			resp := WeblogicAPIResponse{}
			json.Unmarshal([]byte(testCase.apiResponse), &resp)
		}
	}
}

func TestCreateMetrics(t *testing.T) {
	for _, tc := range metricTestCases {
		gauges := make([]prometheus.Gauge, len(tc.metrics))
		for i, ms := range tc.metrics {
			g := prometheus.NewGauge(prometheus.GaugeOpts{
				Name:        ms.name,
				ConstLabels: ms.labels,
			})
			g.Set(ms.value)
			gauges[i] = g
		}

		e, err := New(tc.queries)
		if err != nil {
			t.Fatalf(err.Error())
		}

		genMetrics, err := e.CreateMetrics(&tc.parsedResponse)

		if !reflect.DeepEqual(gauges, genMetrics) {
			t.Errorf("Want %s\nGot %s\n", gauges, genMetrics)
		}
	}
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}
