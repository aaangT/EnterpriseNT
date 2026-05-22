package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Definimos un modelo de datos para pacientes
type NT01 struct {
	C01 string `bson:"_id,omitempty" json:"id"`
	C02 string `bson:"lab21c2" json:"c02"`
	C03 string `bson:"lab21c3" json:"c03"`
}

// JSON plano para pruebas / estudios desde lab57_2026
type TestFlat struct {
	ID        string    `bson:"id" json:"id"`
	Order     int64     `bson:"order" json:"order"`
	Type      string    `bson:"type" json:"type"`
	Service   string    `bson:"service" json:"service"`
	Customer  string    `bson:"customer" json:"customer"`
	Area      string    `bson:"area" json:"area"`
	TestName  string    `bson:"testName" json:"testName"`
	TestID    string    `bson:"testID" json:"testID"`
	Created   time.Time `bson:"created" json:"created"`
	Validated time.Time `bson:"validated" json:"validated"`
	OpTime    float64   `bson:"-" json:"opTime"`
	Status    string    `bson:"status" json:"status"`
}

var client *mongo.Client

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

	router.HandleFunc("/pacientes", getPacientes).Methods("GET")
	router.HandleFunc("/orders", getOrders).Methods("GET")

	log.Println("Servidor corriendo en http://0.0.0.0:8000")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", router))
}

// Funcion para traer pacientes
func getPacientes(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	NTP := []NT01{}

	coleccion := client.Database("EnterpriseNT").Collection("lab21")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := coleccion.Find(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var PA NT01

		err := cursor.Decode(&PA)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		NTP = append(NTP, PA)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(NTP)
}

// Funcion para traer estudios / pruebas desde lab57_2026
func getOrders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tests := []TestFlat{}

	coleccion := client.Database("EnterpriseNT").Collection("lab57_2026")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		bson.D{
			{
				Key: "$lookup",
				Value: bson.D{
					{Key: "from", Value: "lab22_2026"},
					{Key: "localField", Value: "lab22c1"},
					{Key: "foreignField", Value: "lab22c1"},
					{Key: "as", Value: "lab22_docs"},
				},
			},
		},
		bson.D{
			{
				Key: "$project",
				Value: bson.D{
					{Key: "_id", Value: 0},
					{
						Key: "id",
						Value: bson.D{
							{Key: "$toString", Value: "$_id"},
						},
					},
					{Key: "order", Value: "$lab22c1"},
					{Key: "area", Value: "$lab39.lab43.lab43c4"},
					{Key: "testName", Value: "$lab39.lab39c4"},
					{Key: "testID", Value: "$lab39.lab39c2"},
					{Key: "created", Value: "$createdAt"},
					{Key: "validated", Value: "$lab57c18"},
					{
						Key: "status",
						Value: bson.D{
							{Key: "$literal", Value: "validated"},
						},
					},
					{
						Key: "service",
						Value: bson.D{
							{Key: "$arrayElemAt", Value: bson.A{"$lab22_docs.lab10.lab10c2", 0}},
						},
					},
					{
						Key: "type",
						Value: bson.D{
							{Key: "$arrayElemAt", Value: bson.A{"$lab22_docs.lab103.lab103c3", 0}},
						},
					},
					{
						Key: "customer",
						Value: bson.D{
							{Key: "$arrayElemAt", Value: bson.A{"$lab22_docs.lab14.lab14c3", 0}},
						},
					},
				},
			},
		},
	}

	cursor, err := coleccion.Aggregate(ctx, pipeline)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var test TestFlat

		err := cursor.Decode(&test)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		test.Area = normalizarArea(test.Area)
		test.OpTime = calcularOpTime(test.Created, test.Validated)

		tests = append(tests, test)
	}

	if err := cursor.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tests)
}

func normalizarArea(valor string) string {
	valor = strings.TrimSpace(valor)
	valor = strings.ToUpper(valor)

	replacer := strings.NewReplacer(
		"Á", "A",
		"É", "E",
		"Í", "I",
		"Ó", "O",
		"Ú", "U",
		"Ü", "U",
		"Ñ", "N",
	)

	valor = replacer.Replace(valor)

	runes := []rune(valor)

	if len(runes) >= 3 {
		return string(runes[:3])
	}

	return valor
}

func calcularOpTime(created time.Time, validated time.Time) float64 {
	if created.IsZero() || validated.IsZero() {
		return 0
	}

	diferencia := validated.Sub(created)

	if diferencia < 0 {
		return 0
	}

	horas := diferencia.Hours()

	return math.Round(horas*100) / 100
}



lab57c8:
	0 → ordenada
	1 → repetición
	2 → con resultado
	4 → validada
	6 → validada