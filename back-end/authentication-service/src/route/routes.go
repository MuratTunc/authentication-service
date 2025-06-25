package route

import (
	"authentication-service/src/config"
	"authentication-service/src/handler"
	"authentication-service/src/middleware"
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

type Routes struct {
	Config     *config.Config
	Handler    *handler.Handler
	Middleware *middleware.Middleware
}

func NewRoutes(cfg *config.Config, db *sql.DB) *Routes {
	// Create the handler by passing cfg
	h := handler.NewHandler(cfg)
	mw := middleware.NewMiddleware(cfg)
	return &Routes{
		Config:     cfg,
		Handler:    h,
		Middleware: mw,
	}
}

func (r *Routes) Routes() http.Handler {
	mux := chi.NewRouter()

	// Setup global middleware
	r.Middleware.SetupMiddleware(mux)

	// Register public routes
	r.publicRoutes(mux)
	r.Config.Logger.Info("✅ Routes endpoints initialized successfully")

	return mux
}

func (r *Routes) publicRoutes(mux *chi.Mux) {
	mux.Route("/api/v1/auth", func(mux chi.Router) {
		// GET Requests
		mux.Get("/health", r.Handler.HealthCheckHandler)
		mux.Get("/last-user", r.Handler.GetLastUserHandler)
		mux.Get("/get-user-by-mail", r.Handler.GetUserByMailAddressHandler)
		mux.Get("/list-users", r.Handler.ListUsersHandler)

		// PUT Requests
		mux.Put("/update-user", r.Handler.UpdateUserHandler)

		// POST Requests with custom rate limits
		mux.With(httprate.LimitByIP(r.Config.RateLimitRegister, r.Config.RateLimitWindowMinutes*time.Minute)).
			Post("/register-user", r.Handler.RegisterUserHandler)

		mux.With(httprate.LimitByIP(r.Config.RateLimitLogin, r.Config.RateLimitWindowMinutes*time.Minute)).
			Post("/login", r.Handler.LoginUserHandler)

		mux.With(httprate.LimitByIP(r.Config.RateLimitResetCode, r.Config.RateLimitWindowMinutes*time.Minute)).
			Post("/send-mail-reset-code", r.Handler.SendMailResetCodeHandler)

		mux.With(httprate.LimitByIP(r.Config.RateLimitResetPassword, r.Config.RateLimitWindowMinutes*time.Minute)).
			Post("/reset-password", r.Handler.ResetPasswordHandler)

		mux.With(httprate.LimitByIP(r.Config.RateLimitResetCode, r.Config.RateLimitWindowMinutes*time.Minute)).
			Post("/generate-auth-code", r.Handler.GenerateAuthCodeHandler)

		mux.With(httprate.LimitByIP(r.Config.RateLimitResetCode, r.Config.RateLimitWindowMinutes*time.Minute)).
			Post("/verify-auth-code", r.Handler.VerifyAuthCodeHandler)

		mux.Post("/logout", r.Handler.LogoutUserHandler)
		mux.Post("/refresh-jwt-token", r.Handler.RefreshTokenHandler)
		mux.Post("/change-password", r.Handler.ChangePasswordHandler)
		mux.Post("/deactivate-user", r.Handler.DeactivateUserHandler)
		mux.Post("/reactivate-user", r.Handler.ReactivateUserHandler)
		mux.Post("/check-mail-exists", r.Handler.CheckMailExistHandler)
		mux.Post("/verify-mail-reset-code", r.Handler.VerifyResetCodeHandler)

		// DELETE Requests
		mux.Delete("/delete-user", r.Handler.DeleteUserHandler)
	})
}
