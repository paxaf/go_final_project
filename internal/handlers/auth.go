package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

type loginRequest struct {
	Password string `json:"password"`
}

func Login(pass, secret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req loginRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			respondWithError(w, "Ошибка запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()
		pass := os.Getenv("TODO_PASSWORD")
		if req.Password != pass {
			respondWithError(w, "Ошибка авторизации", http.StatusUnauthorized)
			return
		}
		hash := sha256.Sum256([]byte(pass))
		hashString := hex.EncodeToString(hash[:])
		claims := jwt.RegisteredClaims{
			Subject: hashString,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
		if err != nil {
			respondWithError(w, "ошибка подписи токена", http.StatusInternalServerError)
			return
		}
		respondWithJSON(w, http.StatusAccepted, map[string]string{"token": tokenString})

	}
}

func Auth(pass, secret string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if len(pass) == 0 {
				next.ServeHTTP(w, r)
				return
			}
			cookie, err := r.Cookie("token")
			if err != nil {
				respondWithError(w, "Ошибка авторизации", http.StatusUnauthorized)
				return
			}
			jwtToken := cookie.Value
			parsedToken, err := jwt.ParseWithClaims(jwtToken, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})
			if err != nil || !parsedToken.Valid {
				respondWithError(w, "Неверный токен", http.StatusUnauthorized)
				return
			}
			claims, ok := parsedToken.Claims.(*jwt.RegisteredClaims)
			if !ok {
				respondWithError(w, "Ошибка чтения токена", http.StatusUnauthorized)
				return
			}
			hash := sha256.Sum256([]byte(pass))
			expectedHash := hex.EncodeToString(hash[:])
			if claims.Subject != expectedHash {
				respondWithError(w, "Неверные учетные данные", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
