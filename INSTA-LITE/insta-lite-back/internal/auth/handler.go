package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/huguescodeur/insta-lite/internal/middlewares"
	"github.com/huguescodeur/insta-lite/internal/pkg/responses"
	"github.com/huguescodeur/insta-lite/internal/user"
)

type AuthHandler struct {
	service  *AuthService
	validate *validator.Validate
}

func NewHandlerAuth(s *AuthService) *AuthHandler {
	return &AuthHandler{
		service:  s,
		validate: validator.New(),
	}
}

// RegisterHandler godoc
// @Summary      Inscription d'un nouvel utilisateur
// @Description  Crée un compte et retourne le token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      RegisterRequest  true  "Données d'inscription"
// @Success      201   {object}  map[string]any
// @Failure      400   {string}  string  "Format JSON invalide ou validation échouée"
// @Failure      500   {string}  string  "Erreur interne"
// @Router       /auth/register [post]
func (h *AuthHandler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	u := user.User{
		Username:     req.Username,
		Email:        req.Email,
		FullName:     req.FullName,
		PasswordHash: req.Password,
	}

	user, token, err := h.service.Register(ctx, &u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"user":  user,
		"token": token,
	})

}

// LoginHandler godoc
// @Summary      Connexion
// @Description  Authentifie un utilisateur (email ou username) et retourne le token JWT
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body  body      LoginRequest  true  "Identifiants de connexion"
// @Success      200   {object}  map[string]any
// @Failure      400   {string}  string  "Format JSON invalide ou validation échouée"
// @Failure      401   {string}  string  "Identifiants incorrects"
// @Failure      500   {string}  string  "Erreur interne"
// @Router       /auth/login [post]
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		responses.SendValidationError(w, err)
		return
	}

	user, token, err := h.service.Login(ctx, req.Identifier, req.Password)
	if err != nil {
		if err.Error() == "identifiants incorrect" {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user":  user,
		"token": token,
	})

}

func (h *AuthHandler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Format de token invalide (Bearer requis)", http.StatusBadRequest)
		return
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Le handler appelle simplement le service sans lui passer Redis en paramètre !
	err := h.service.Logout(ctx, tokenString)
	if err != nil {
		http.Error(w, "Impossible de procéder à la déconnexion : "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Déconnexion réussie"}`))
}

func (h *AuthHandler) AuthRoutes() chi.Router {
	r := chi.NewRouter()

	registerLimiter := middlewares.ConfigurableRateLimiter("auth_register", 2, time.Minute)
	loginLimiter := middlewares.ConfigurableRateLimiter("auth_login", 3, time.Minute)

	r.With(registerLimiter).Post("/register", h.RegisterHandler)
	r.With(loginLimiter).Post("/login", h.LoginHandler)

	return r
}
