package main

import (
	"errors"
	"fmt"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/time/rate"
	"library/internal/data"
	"library/internal/validation"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
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

func (app *application) enableCors(next http.Handler) http.Handler {
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

func (app *application) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//reqInfo, err := httputil.DumpRequest(r, false)
		//if err != nil {
		//	app.serverErrorResponse(w, r, err)
		//	return
		//}
		//app.logger.PrintInfo(string(reqInfo), nil)

		properties := map[string]string{
			"address":  r.RemoteAddr,
			"method":   r.Method,
			"uri":      r.URL.RequestURI(),
			"protocol": r.Proto,
		}

		app.logger.PrintInfo("Log request information", properties)

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {

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
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

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

func (app *application) requireActivatedUser(next httprouter.Handle) httprouter.Handle {
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

func (app *application) requireAuthenticatedUser(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, pm httprouter.Params) {
		user := app.ctxGetUser(r)
		if user.IsAnonymous() {
			app.unauthorizedResponse(w, r)
			return
		}

		next(w, r, pm)
	}
}

func (app *application) requirePermission(permission string, next httprouter.Handle) httprouter.Handle {
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

func (app *application) authenticate(next http.Handler) http.Handler {
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
