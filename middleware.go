package negronilogrus

import (
    "fmt"
    "github.com/Sirupsen/logrus"
    "github.com/codegangsta/negroni"
    "github.com/jabong/canonburst/log"
    "net/http"
    "time"
)

// Middleware is a middleware handler that logs the request as it goes in and the response as it goes out.
type Middleware struct {
    // Logger is the log.Logger instance used to log messages with the Logger middleware
    Logger *logrus.Logger
    // Name is the name of the application as recorded in latency metrics
    Name string
}

// NewMiddleware returns a new *Middleware, yay!
func NewMiddleware() *Middleware {
    return NewCustomMiddleware(logrus.InfoLevel, &logrus.TextFormatter{}, "web")
}

// NewCustomMiddleware builds a *Middleware with the given level and formatter
func NewCustomMiddleware(level logrus.Level, formatter logrus.Formatter, name string) *Middleware {
    log := logrus.New()
    log.Level = level
    log.Formatter = formatter

    return &Middleware{Logger: log, Name: name}
}

func (l *Middleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
    start := time.Now()
    l.Logger.WithFields(logrus.Fields{
        "method":  r.Method,
        "request": r.RequestURI,
        "remote":  r.RemoteAddr,
    }).Info("started handling request")

    log.Info(fmt.Sprintf("started handling request: method=%s remote=%s request=%s", r.Method, r.RemoteAddr, r.RequestURI))

    next(rw, r)

    latency := time.Since(start)
    res := rw.(negroni.ResponseWriter)
    l.Logger.WithFields(logrus.Fields{
        "status":      res.Status(),
        "method":      r.Method,
        "request":     r.RequestURI,
        "remote":      r.RemoteAddr,
        "text_status": http.StatusText(res.Status()),
        "took":        latency,
        fmt.Sprintf("measure#%s.latency", l.Name): latency.Nanoseconds(),
    }).Info("completed handling request")

    msg := fmt.Sprintf("completed handling request: measure#%s.latency=%d method=%s remote=%s request=%s status=%d text_status=%s took=%s", l.Name, latency.Nanoseconds(), r.Method, r.RemoteAddr, r.RequestURI, res.Status(), http.StatusText(res.Status()), latency)
    if res.Status() == http.StatusOK {
        log.Info(msg)
    } else {
        log.Crit(msg)
    }
}
