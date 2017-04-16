# Zipkin example

This example starts up a [Zipkin](http://zipkin.io) server and an example web application that send traces to the Zipkin backend. The example application implement the [Opentracing](http://opentracing.io) API.

### Requirements

* A working installation of Docker and Docker Compose

### Run the example
`docker-compose up`

Starts the Zipkin server and API as well as the example application.

Example application can be viewed at: `http://localhost:8080`

Zipkin server can be viewed at `http://localhost:9411`

