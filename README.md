# Opentracing Zipkin example

An example application instrumented with tracing to show how requests flow between components. The [Opentracing](http://opentracing.io) API is implemented to collect trace information and is configured to sent traces to a [Zipkin](http://zipkin.io) server where it can be viewed through the Zipkin UI. 

### Run the example

Start the Zipkin server and example application
`docker-compose up`

Example application can be viewed at: `http://localhost:8080`

Zipkin server can be viewed at `http://localhost:9411`

