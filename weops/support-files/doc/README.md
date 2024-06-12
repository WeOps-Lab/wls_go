## 嘉为蓝鲸weblogic插件使用说明

## 使用说明

### 插件功能
WebLogic监控探针通过使用WebLogic Server的RESTful API ( `/management/weblogic/latest/serverRuntime/search` ) 导出监控指标数据。

### 版本支持

操作系统支持: linux, windows

是否支持arm: 支持

**组件支持版本：**

Weblogic Server: `12.2.1+`

**是否支持远程采集:**

是

### 参数说明


| **参数名**              | **含义**                       | **是否必填** | **使用举例**       |
|----------------------|------------------------------|----------|----------------|
| USERNAME             | weblogic登录用户名(环境变量)          | 是        | weblogic       |
| PASSWORD             | weblogic登录密码(环境变量)           | 是        | welcome        |
| --host               | weblogic服务实例IP地址             | 是        | 127.0.0.1      |
| --port               | weblogic服务实例端口               | 是        | 7001           |
| --config-file        | 采集配置文件(文件参数), 默认内容已填写, 不需要修改 | 是        | config.yaml    |
| --web.listen-address | exporter监听id及端口地址            | 否        | 127.0.0.1:9601 |


### 使用指引
1. 验证登录信息 
   访问Weblogic Servers实例, 登录 `host:port/console` 。默认管理服务器端口一般是7001  
   输入账户和密码, 验证是否账户是否正确。  
   
2. 如果需要采集数据源监控指标  
   [配置数据源指引](https://www.oracle.com/webfolder/technetwork/tutorials/obe/fmw/wls/12c/12_1_3/04/04-ConfigDataSource/configds.html)  
   如果没有配置数据源, 那么将采集不到 `wls_datasource_`类指标。



### 指标简介
| **指标ID**                                        | **指标中文名**                   | **维度ID**                                                     | **维度含义**                       |
|-------------------------------------------------|-----------------------------|--------------------------------------------------------------|--------------------------------|
| weblogic_probe_success                          | weblogic监控探针运行状态            | -                                                            | -                              |
| wls_server_health_health_state                  | weblogic实例健康状态              | server_name                                                  | 服务器名称                          |
| wls_jvm_heap_free_current                       | weblogic JVM当前空闲堆内存         | server_name                                                  | 服务器名称                          |
| wls_jvm_heap_free_percent                       | weblogic JVM堆内存空闲百分比        | server_name                                                  | 服务器名称                          |
| wls_jvm_heap_size_current                       | weblogic JVM当前堆内存大小         | server_name                                                  | 服务器名称                          |
| wls_servlet_execution_time_average              | weblogic Servlet平均执行时间      | application_runtime, component_runtime, server_name, servlet | 应用程序运行时, 组件运行时, 服务器名称, Servlet |
| wls_servlet_execution_time_high                 | weblogic Servlet最大执行时间      | application_runtime, component_runtime, server_name, servlet | 应用程序运行时, 组件运行时, 服务器名称, Servlet |
| wls_servlet_execution_time_total                | weblogic Servlet总执行时间       | application_runtime, component_runtime, server_name, servlet | 应用程序运行时, 组件运行时, 服务器名称, Servlet |
| wls_servlet_invocation_total_count              | weblogic Servlet调用总次数       | application_runtime, component_runtime, server_name, servlet | 应用程序运行时, 组件运行时, 服务器名称, Servlet |
| wls_threadpool_stuck_thread_count               | weblogic线程池卡住线程数            | server_name, threadpool                                      | 服务器名称, 线程池                     |
| wls_threadpool_throughput                       | weblogic线程池吞吐量              | server_name, threadpool                                      | 服务器名称, 线程池                     |
| wls_threadpool_hogging_thread_count             | weblogic线程池占用线程数            | server_name, threadpool                                      | 服务器名称, 线程池                     |
| wls_threadpool_pending_user_request_count       | weblogic线程池挂起用户请求数          | server_name, threadpool                                      | 服务器名称, 线程池                     |
| wls_threadpool_standby_thread_count             | weblogic线程池待机线程数            | server_name, threadpool                                      | 服务器名称, 线程池                     |
| wls_threadpool_queue_length                     | weblogic线程池队列长度             | server_name, threadpool                                      | 服务器名称, 线程池                     |
| wls_threadpool_execute_thread_idle_count        | weblogic线程池空闲执行线程数          | server_name, threadpool                                      | 服务器名称, 线程池                     |
| wls_threadpool_execute_thread_total_count       | weblogic线程池执行线程总数           | server_name, threadpool                                      | 服务器名称, 线程池                     |
| wls_jms_connections_current_count               | weblogic JMS当前连接数           | JMS, server_name                                             | 消息服务名称, 服务器名称                  |
| wls_jms_connections_high_count                  | weblogic JMS最大连接数           | JMS, server_name                                             | 消息服务名称, 服务器名称                  |
| wls_jms_connections_total_count                 | weblogic JMS连接总数            | JMS, server_name                                             | 消息服务名称, 服务器名称                  |
| wls_webapp_deployment_state                     | weblogic Web应用部署状态          | application_runtime, component_runtime, server_name          | 应用程序运行时, 组件运行时, 服务器名称          |
| wls_webapp_open_sessions_current_count          | weblogic Web应用当前打开会话数       | application_runtime, component_runtime, server_name          | 应用程序运行时, 组件运行时, 服务器            |
| wls_webapp_open_sessions_high_count             | weblogic Web应用最大打开会话数       | application_runtime, component_runtime, server_name          | 应用程序运行时, 组件运行时, 服务器            |
| wls_webapp_servlet_reload_check_secs            | weblogic Web应用Servlet重载检查秒数 | application_runtime, component_runtime, server_name          | 应用程序运行时, 组件运行时, 服务器            |
| wls_webapp_sessions_opened_total_count          | weblogic Web应用打开的会话总数       | application_runtime, component_runtime, server_name          | 应用程序运行时, 组件运行时, 服务器            |
| wls_datasource_connections_total_count          | weblogic数据源连接总数             | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_curr_capacity                    | weblogic数据源当前容量             | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_deployment_state                 | weblogic数据源部署状态             | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_failures_to_reconnect_count      | weblogic数据源失败重连总数           | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_active_connections_current_count | weblogic数据源当前连接总数           | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_num_available                    | weblogic数据源可用连接数            | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_num_available                    | weblogic数据源不可用连接数           | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_waiting_for_connection_total     | weblogic数据源连接等待总数           | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_connection_delay_time            | weblogic数据源连接延时             | datasource, jdbc_service, server_name                        | 数据源名称, jdbc服务名称, 服务器名称         |
| wls_datasource_state                            | weblogic数据源运行状态             | datasource, jdbc_service, server_name, state                 | 数据源名称, jdbc服务名称, 服务器名称, 运行状态   |


### 版本日志

#### weops_weblogic_exporter 2.1.1

- weops调整


添加“小嘉”微信即可获取weblogic监控指标最佳实践礼包，其他更多问题欢迎咨询

<img src="https://wedoc.canway.net/imgs/img/小嘉.jpg" width="50%" height="50%">
