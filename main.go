package main

import (
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/openzipkin/zipkin-go-opentracing"
	"log"
	"math/rand"
	"net/http"
	"time"
)

// Acts as our index page
func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<a href="/home"> Click here to start a request </a>`))
}
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Request started"))

	span := opentracing.StartSpan("/home")
	defer span.Finish()

	asyncReq, _ := http.NewRequest("GET", "http://localhost:8080/async", nil)
	err := span.Tracer().Inject(span.Context(), opentracing.TextMap, opentracing.HTTPHeadersCarrier(asyncReq.Header))
	if err != nil {
		log.Fatalf("Could not inject span context into header: %v", err)
	}

	go func() {
		if _, err := http.DefaultClient.Do(asyncReq); err != nil {
			ext.Error.Set(span, true)
			span.LogKV("GET async error", err)
		}
	}()
	srcReq, _ := http.NewRequest("GET", "http://localhost:8080/service", nil)
	err = span.Tracer().Inject(span.Context(), opentracing.TextMap, opentracing.HTTPHeadersCarrier(srcReq.Header))
	if err != nil {
		log.Fatalf("Could not inject span context into header: %v", err)
	}

	if _, err := http.DefaultClient.Do(srcReq); err != nil {
		ext.Error.Set(span, true)
		span.LogKV("GET service error", err)
	}

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	w.Write([]byte("Request done!"))
}

// Mocks a service endpoint that makes a DB call
func serviceHandler(w http.ResponseWriter, r *http.Request) {
	var span opentracing.Span
	opName := r.URL.Path

	wireContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		span = opentracing.StartSpan(opName)
	} else {
		span = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
	}
	defer span.Finish()

	dbReq, _ := http.NewRequest("GET", "http://localhost:8080/db", nil)
	err = span.Tracer().Inject(span.Context(), opentracing.TextMap, opentracing.HTTPHeadersCarrier(dbReq.Header))
	if err != nil {
		log.Fatalf("Could not inject span context into header: %v", err)
	}
	if _, err = http.DefaultClient.Do(dbReq); err != nil {
		ext.Error.Set(span, true)
		span.LogKV("GET db error", err)
	}

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	// ...
}

// Mocks a DB call
func dbHandler(w http.ResponseWriter, r *http.Request) {
	var sp opentracing.Span
	opName := r.URL.Path
	wireContext, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil {
		sp = opentracing.StartSpan(opName)
	} else {
		sp = opentracing.StartSpan(opName, opentracing.ChildOf(wireContext))
	}
	defer sp.Finish()

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)
	// here would be the actual call to a DB.
}

func main() {
	port := 8080
	addr := fmt.Sprintf(":%d", port)

	zipkinCollector, err := zipkintracer.NewHTTPCollector("http://zipkin:9411/api/v1/spans")
	if err != nil {
		log.Fatalf("unable to create Zipkin HTTP collector: %+v", err)
	}
	defer zipkinCollector.Close()

	zipkinRecorder := zipkintracer.NewRecorder(zipkinCollector, false, "0.0.0.0:8080", "payprocessor")
	zipkinTracer, err := zipkintracer.NewTracer(zipkinRecorder, zipkintracer.ClientServerSameSpan(true), zipkintracer.TraceID128Bit(true))
	if err != nil {
		log.Fatalf("unable to create Zipkin tracer: %+v", err)
	}

	opentracing.InitGlobalTracer(zipkinTracer)

	mux := http.NewServeMux()
	mux.HandleFunc("/", indexHandler)
	mux.HandleFunc("/home", homeHandler)
	mux.HandleFunc("/async", serviceHandler)
	mux.HandleFunc("/service", serviceHandler)
	mux.HandleFunc("/db", dbHandler)
	fmt.Printf("Go to http://localhost:%d/home to start a request!\n", port)
	log.Fatal(http.ListenAndServe(addr, mux))
}
