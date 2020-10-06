package trace

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"github.com/gin-gonic/gin"
	"go.opencensus.io/exporter/stackdriver/propagation"
	"go.opencensus.io/plugin/ochttp"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"
	"go.opencensus.io/trace"
	oppropagation "go.opencensus.io/trace/propagation"
)

// Setter defines a context that enables setting values.
type Setter interface {
	Set(string, interface{})
}

const key = "tracespan"

func FromContext(ctx context.Context) *trace.Span {
	if span := ctx.Value(key); span == nil {
		return nil
	} else {
		return span.(*trace.Span)
	}
}

func ToContext(s Setter, span *trace.Span) {
	s.Set(key, span)
}

func NewSpan(ctx context.Context, spanName string) *trace.Span {
	span := FromContext(ctx)
	span.SetName(spanName)
	return span
}

func Trace(traceClient trace.Exporter) gin.HandlerFunc {
	exporter, err := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: os.Getenv("GOOGLE_CLOUD_PROJECT"),
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(exporter)

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		parts := strings.Split(path, "/")
		var handler string
		if len(parts) > 2 {
			handler = strings.Join(parts[:3], "/")
		}

		if handler == "" {
			handler = "/"
		}

		//span := traceClient.NewSpan(handler)
		//ToContext(c, span)
		//c.Next()
		//span.Finish()
	}
}

func Middleware(service string) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func TraceMiddleware(c *gin.Context) {
	h := &handler{
		propagation:      &propagation.HTTPFormat{},
		IsPublicEndpoint: false,
		StartOptions: trace.StartOptions{
			SpanKind: trace.SpanKindServer,
		},
	}

	h.HandlerFunc(c)
}

type handler struct {
	name             string
	propagation      oppropagation.HTTPFormat
	StartOptions     trace.StartOptions
	IsPublicEndpoint bool
}

func (h *handler) HandlerFunc(c *gin.Context) {
	path := c.Request.URL.Path

	parts := strings.Split(path, "/")
	var handler string
	if len(parts) > 2 {
		handler = strings.Join(parts[:3], "/")
	}

	if handler == "" {
		handler = "/"
	}

	h.name = "/spike" + handler
	var traceEnd, statsEnd func()
	c.Request, traceEnd = h.startTrace(c.Writer, c.Request)
	c.Writer, statsEnd = h.startStats(c.Writer, c.Request)

	c.Next()

	statsEnd()
	traceEnd()
}

func (h *handler) startTrace(_ gin.ResponseWriter, r *http.Request) (*http.Request, func()) {
	ctx := r.Context()
	var span *trace.Span
	sc, ok := h.extractSpanContext(r)
	if ok && !h.IsPublicEndpoint {
		ctx, span = trace.StartSpanWithRemoteParent(ctx, h.name, sc, trace.WithSampler(h.StartOptions.Sampler), trace.WithSpanKind(h.StartOptions.SpanKind))
	} else {
		ctx, span = trace.StartSpan(ctx, h.name, trace.WithSampler(h.StartOptions.Sampler), trace.WithSpanKind(h.StartOptions.SpanKind))
		if ok {
			span.AddLink(trace.Link{
				TraceID:    sc.TraceID,
				SpanID:     sc.SpanID,
				Type:       trace.LinkTypeChild,
				Attributes: nil,
			})
		}
	}
	span.AddAttributes(requestAttrs(r)...)
	return r.WithContext(ctx), span.End
}

func (h *handler) extractSpanContext(r *http.Request) (trace.SpanContext, bool) {
	return h.propagation.SpanContextFromRequest(r)
}

func (h *handler) startStats(w gin.ResponseWriter, r *http.Request) (gin.ResponseWriter, func()) {
	ctx, _ := tag.New(r.Context(),
		tag.Upsert(ochttp.Host, r.URL.Host),
		tag.Upsert(ochttp.Path, r.URL.Path),
		tag.Upsert(ochttp.Method, r.Method))
	track := &trackingResponseWriter{
		start:          time.Now(),
		ctx:            ctx,
		ResponseWriter: w,
	}
	if r.Body == nil {
		// TODO: Handle cases where ContentLength is not set.
		track.reqSize = -1
	} else if r.ContentLength > 0 {
		track.reqSize = r.ContentLength
	}
	stats.Record(ctx, ochttp.ServerRequestCount.M(1))
	return track, track.end
}

func requestAttrs(r *http.Request) []trace.Attribute {
	return []trace.Attribute{
		trace.StringAttribute(ochttp.PathAttribute, r.URL.Path),
		trace.StringAttribute(ochttp.HostAttribute, r.URL.Host),
		trace.StringAttribute(ochttp.MethodAttribute, r.Method),
		trace.StringAttribute(ochttp.UserAgentAttribute, r.UserAgent()),
	}
}

func responseAttrs(resp *http.Response) []trace.Attribute {
	return []trace.Attribute{
		trace.Int64Attribute(ochttp.StatusCodeAttribute, int64(resp.StatusCode)),
	}
}

type trackingResponseWriter struct {
	gin.ResponseWriter
	ctx     context.Context
	reqSize int64
	start   time.Time
	endOnce sync.Once
}

var _ gin.ResponseWriter = (*trackingResponseWriter)(nil)

func (t *trackingResponseWriter) end() {
	t.endOnce.Do(func() {
		m := []stats.Measurement{
			ochttp.ServerLatency.M(float64(time.Since(t.start)) / float64(time.Millisecond)),
			ochttp.ServerResponseBytes.M(int64(t.Size())),
		}
		if t.reqSize >= 0 {
			m = append(m, ochttp.ServerRequestBytes.M(t.reqSize))
		}
		status := t.Status()
		if status == 0 {
			status = http.StatusOK
		}
		ctx, _ := tag.New(t.ctx, tag.Upsert(ochttp.StatusCode, strconv.Itoa(status)))
		stats.Record(ctx, m...)
	})
}
