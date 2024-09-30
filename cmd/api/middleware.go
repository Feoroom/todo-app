package main

import (
	"errors"
	"expvar"
	"fmt"
	"github.com/felixge/httpsnoop"
	"github.com/julienschmidt/httprouter"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
	"library/internal/data"
	"library/internal/validation"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

func (app *Application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *Application) enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range app.config.CORS.AllowedOrigins {
				if origin == app.config.CORS.AllowedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if isPreflight(r) {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

						w.WriteHeader(http.StatusOK)
						return
					}
					break
				}
			}
		}

		//w.Header().Set("Access-Control-Allow-Origin", "*")

		next.ServeHTTP(w, r)
	})
}

func (app *Application) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//reqInfo, err := httputil.DumpRequest(r, false)
		//if err != nil {
		//	app.serverErrorResponse(w, r, err)
		//	return
		//}
		//app.logger.PrintInfo(string(reqInfo), nil)

		app.logger.Info("Log request information",
			slog.String("address", realip.FromRequest(r)),
			slog.String("method", r.Method),
			slog.String("uri", r.URL.RequestURI()),
			slog.String("protocol", r.Proto),
		)

		next.ServeHTTP(w, r)
	})
}

func (app *Application) metrics(next http.Handler) http.Handler {
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTime := expvar.NewInt("total_processing_time")
	totalResponsesByStatus := expvar.NewMap("responses_status_codes")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		totalRequestsReceived.Add(1)

		metrics := httpsnoop.CaptureMetrics(next, w, r)

		totalResponsesSent.Add(1)
		totalProcessingTime.Add(metrics.Duration.Microseconds())
		totalResponsesByStatus.Add(strconv.Itoa(metrics.Code), 1)

	})
}

func (app *Application) rateLimit(next http.Handler) http.Handler {

	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		time.Sleep(time.Minute)

		mu.Lock()

		for ip, client := range clients {
			if time.Since(client.lastSeen) > 3*time.Minute {
				delete(clients, ip)
			}
		}

		mu.Unlock()
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if app.config.Limiter.Enabled {
			//ip, _, err := net.SplitHostPort(r.RemoteAddr)
			//if err != nil {
			//	app.serverErrorResponse(w, r, err)
			//	return
			//}
			ip := realip.FromRequest(r)

			mu.Lock()

			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(
					rate.Limit(app.config.Limiter.Rps),
					app.config.Limiter.Burst)}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			mu.Unlock()
		}

		next.ServeHTTP(w, r)
	})
}

func (app *Application) requireActivatedUser(next httprouter.Handle) httprouter.Handle {
	fn := func(w http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		user := app.ctxGetUser(r)
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next(w, r, pm)
	}

	return app.requireAuthenticatedUser(fn)
}

func (app *Application) requireAuthenticatedUser(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		user := app.ctxGetUser(r)
		if user.IsAnonymous() {
			app.unauthorizedResponse(w, r)
			return
		}

		next(w, r, pm)
	}
}

func (app *Application) requirePermission(permission string, next httprouter.Handle) httprouter.Handle {
	fn := func(w http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		user := app.ctxGetUser(r)

		perm, err := app.models.Permissions.GetForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if !perm.Include(permission) {
			app.notPermittedResponse(w, r)
			return
		}

		next(w, r, pm)
	}
	return app.requireAuthenticatedUser(fn)
}

func (app *Application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Authorization")

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			r = app.ctxSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		authParts := strings.Split(authHeader, " ")
		if len(authParts) != 2 || authParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
		}

		token := authParts[1]

		v := validation.New()

		if data.ValidateTokenPlainText(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.ctxSetUser(r, user)

		next.ServeHTTP(w, r)
	})
}
