# wls_go
A Prometheus exporter for Weblogic written in Go

The exporter runs as a standalone web service driven by a yaml config file. Once you have configured the application, scrape it on /probe, providing a host and port parameter that tell it where to reach the intended Weblogic server, and credentials using HTTP basic auth. It will then probe the Weblogic API, convert the response into metrics as per your configuration, and return them. For example, using curl:

`curl -u 'Weblogic:Welcome123' http://localhost:9325/probe?host=weblogic.mydomain.io&port=7100`

# Getting Started
The exporter comes with a spec file for building an RPM which you can pass to rpmbuild. Otherwise you can simply clone the repo and `go build -o weblogic_exporter src/main.go`.

### Configuring
By default, the exporter attempts to load config.yaml in its working directory. You can pass in the `--config-file` flag
to change that behaviour. 

The configuration file format looks like the following:
```yaml
---
listen_port: 8443
tls_cert_path: server.crt
tls_key_path: server.key
mbeans:
    label_name: server
    label_value_attribute: name
    children:
        JVMRuntime:
            metric_prefix: wls_jvm_
            fields: [heapFreeCurrent, heapFreePercent, heapSizeCurrent]
        applicationRuntimes:
            label_name: application_runtime
            label_value_attribute: name
            children:
                componentRuntimes:
                    metric_prefix: wls_webapp_
                    label_name: component_runtime
                    label_value_attribute: name
                    fields: [deploymentState, sessionsOpenedTotalCount ]
                    children:
                        servlets:
                          metric_prefix: wls_servlet_
                          label_name: servlet
                          label_value_attribute: servletName
                          fields: [ invocationTotalCount, executionTimeAverage, executionTimeHigh, executionTimeTotal ]
        JDBCServiceRuntime:
            label_name: jdbc_service
            label_value_attribute: name
            children:
                JDBCDataSourceRuntimeMBeans:
                    metric_prefix: wls_datasource_
                    label_value_attribute: name
                    label_name: datasource
                    fields: [ connectionsTotalCount, deploymentState, currCapacity ]
                    string_fields:
                        - name: state
                          value_set:
                            - Running
                            - Suspended
                            - Shutdown
                            - Overloaded
                            - Unknown

```
The top level configration allows the following:
`
* `listen_port` - Integer. Which port the exporter should listen on. By default this is 9325.
* `tls_cert_path` - String. The path to the TLS certificate used when the exporter listens via TLS. Must include the entire CA chain as well as the server cert, appended together in PEM format. 
* `tls_key_path` - String. The TLS private key to use. 
* `mbeans`: - Map/Dict. Configuration for which MBeans to expose. See [Selecting which MBeans and Attributes to Return](#Selecting-which-MBeans-and-Attributes-to-Return)

If neither `tls_cert_path` nor `tls_key_path` are present, the server will listen on plain HTTP.

Essentially, the configuration mimics the Weblogic MBean tree, beginning at the serverRuntime MBean which is the root of 
Weblogic runtime MBean tree. You can find more about MBeans [here](https://docs.oracle.com/middleware/1221/wls/WLMBR/core/index.html). 

### Selecting which MBeans and Attributes to Return
MBeans are exposed by listing them as under the `children` section of a parent MBean. For example, in the config above you can see that the `JVMRuntime` MBean is listed a child of the root MBean (which is `ServerRuntime`). If you open up the MBean reference above for `ServerRuntime`, you can see in the *Related MBeans* section that `JVMRuntime` is a child of `ServerRuntime`. 

Underneath the MBean definition, you may specify the following fields:
* `label_name` - String. This is the name of the label that will end up in your Prometheus metric.
* `label_value_attribute` - String. This is the attribute of the MBean the exporter will use to populate the label value to match the label name you've selected. For example, you may use the label_name `datasource` for a JDBCDataSourceRuntimeMBean, and the `name` attribute that identifies the datasource. 
* `fields` - Array. These are attributes you wish to return as metrics. Note that these must return numerical values, or they will be ignored. Weblogic's API tends to be relatively inconsistent with what it returns here, but you can see what is returned in the reference. You may also specify the healthState attribute here, even though its not numerical. This is because the healthState response is fairly complicated, so the exporter is hardcoded to identify and handle it appropriately. 
* `string_fields` - Array. These are attributes that return strings that you might want to expose as metrics. Be careful, this is not intended to expose arbitrary string like exceptions or error messages, only attributes with a known set of values like deployment states and health states. Each entry must contain:
  * `name`: String. The name of the attribute
  * `value_set`: Array of strings. Represents all the possible values that may be returned. The exporter will create metrics for all of them, with a value of 0. Only the active state retuned in the response will have a value of 1. ** Note ** If you leave a state off this list, and it is returned by the API, it will be silently ignored. You ** must * enumerate all possible states here for accurate metrics. Often, the MBean reference will tell you all the possible states.
* `children`: Map/Dict. Child MBeans.