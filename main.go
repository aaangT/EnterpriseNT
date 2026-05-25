package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

type NT01 struct {
	C01 string `bson:"_id,omitempty"`
	C02 string `bson:"lab21c2"`
	C03 string `bson:"lab21c3"`
}

var client *mongo.Client

var jwtSecret = []byte("CAMBIA_ESTA_CLAVE_SECRETA_SUPER_SEGURA")

type LoginRequest struct {
	Username string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Authenticated bool          `json:"authenticated"`
	Token         string        `json:"token,omitempty"`
	ExpiresIn     int64         `json:"expiresIn,omitempty"`
	User          LoginUserData `json:"user,omitempty"`
	Message       string        `json:"message,omitempty"`
}

type LoginUserData struct {
	ID       string `json:"id"`
	Username string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
}

type AppUser struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	Username     string             `bson:"email"`
	Name         string             `bson:"name"`
	PasswordHash string             `bson:"passwordHash"`
	Role         string             `bson:"role"`
	Active       bool               `bson:"active"`
	Created      time.Time          `bson:"created"`
	LastLogin    *time.Time         `bson:"lastLogin"`
}

type AppClaims struct {
	UserID   string `json:"userId"`
	Username string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("Error conectando a MongoDB:", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("No se pudo hacer ping a MongoDB:", err)
	}

	router := mux.NewRouter()

	//router.HandleFunc("/pacientes", getPacientes).Methods("GET")
	//router.HandleFunc("/orders", getOrdersWithTests).Methods("GET")
	//router.HandleFunc("/orderst", getOrdersWithTests).Methods("GET")
	router.HandleFunc("/login", login).Methods("POST")
	router.HandleFunc("/orders", authMiddleware(getOrdersWithTests)).Methods("GET")
	router.HandleFunc("/paciente/{id}", authMiddleware(getPaciente)).Methods("GET")

	log.Println("Servidor corriendo en http://0.0.0.0:8000")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", enableCORS(router)))
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var loginData LoginRequest

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Authenticated: false,
			Message:       "Datos inválidos",
		})
		return
	}

	if loginData.Username == "" || loginData.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LoginResponse{
			Authenticated: false,
			Message:       "Usuario y contraseña son requeridos",
		})
		return
	}

	collection := client.Database("EnterpriseNT").Collection("app_users")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user AppUser

	err = collection.FindOne(ctx, bson.M{
		"email":  loginData.Username,
		"active": true,
	}).Decode(&user)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Authenticated: false,
			Message:       "Usuario o contraseña incorrectos",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginData.Password))
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Authenticated: false,
			Message:       "Usuario o contraseña incorrectos",
		})
		return
	}

	expirationTime := time.Now().Add(8 * time.Hour)

	claims := AppClaims{
		UserID:   user.ID.Hex(),
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "EnterpriseNT-API",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LoginResponse{
			Authenticated: false,
			Message:       "Error generando token",
		})
		return
	}

	now := time.Now()

	_, _ = collection.UpdateOne(
		ctx,
		bson.M{"_id": user.ID},
		bson.M{
			"$set": bson.M{
				"lastLogin": now,
			},
		},
	)

	json.NewEncoder(w).Encode(LoginResponse{
		Authenticated: true,
		Token:         tokenString,
		ExpiresIn:     int64(8 * 60 * 60),
		User: LoginUserData{
			ID:       user.ID.Hex(),
			Username: user.Username,
			Name:     user.Name,
			Role:     user.Role,
		},
	})
}

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(bson.M{
				"authenticated": false,
				"message":       "Token requerido",
			})
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(bson.M{
				"authenticated": false,
				"message":       "Formato de token inválido",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &AppClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(bson.M{
				"authenticated": false,
				"message":       "Token inválido o expirado",
			})
			return
		}

		userObjectID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(bson.M{
				"authenticated": false,
				"message":       "Usuario inválido",
			})
			return
		}

		collection := client.Database("EnterpriseNT").Collection("app_users")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var user AppUser

		err = collection.FindOne(ctx, bson.M{
			"_id":    userObjectID,
			"active": true,
		}).Decode(&user)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(bson.M{
				"authenticated": false,
				"message":       "Usuario inactivo o no autorizado",
			})
			return
		}

		next(w, r)
	}
}

func getPaciente(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id := params["id"]

	coleccion := client.Database("EnterpriseNT").Collection("lab21")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var paciente NT01

	err := coleccion.FindOne(ctx, bson.M{"lab21c2": id}).Decode(&paciente)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Paciente no encontrado", http.StatusNotFound)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(paciente)
}

func getOrdersWithTests(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database("EnterpriseNT").Collection("lab22_2026")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$lookup", Value: bson.M{
			"from":         "lab57_2026",
			"localField":   "lab22c1",
			"foreignField": "lab22c1",
			"as":           "tests",
		}}},

		{{Key: "$project", Value: bson.M{
			"_id":       0,
			"order":     "$lab22c1",
			"service":   "$lab10.lab10c2",
			"type":      "$lab103.lab103c2",
			"customer":  "$lab14.lab14c3",
			"createdAt": "$createdAt",

			"status": getOrderStatusExpression(),
			"opTime": getRandomOpTimeExpression(),
			"tests":  getTestsExpression(),
		}}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		http.Error(w, "Error ejecutando aggregate: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	var results []bson.M

	if err := cursor.All(ctx, &results); err != nil {
		http.Error(w, "Error leyendo resultados: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(results)
}
