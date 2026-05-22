package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NT01 struct {
	C01 string `bson:"_id,omitempty"`
	C02 string `bson:"lab21c2"`
	C03 string `bson:"lab21c3"`
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

	//router.HandleFunc("/pacientes", getPacientes).Methods("GET")
	//router.HandleFunc("/orders", getOrders).Methods("GET")
	router.HandleFunc("/orders", getOrdersWithTests).Methods("GET")
	router.HandleFunc("/paciente/{id}", getPaciente).Methods("GET")

	log.Println("Servidor corriendo en http://0.0.0.0:8000")
	log.Fatal(http.ListenAndServe("0.0.0.0:8000", router))
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
			"order":     "$lab22c1",
			"service":   "$lab10.lab10c2",
			"type":      "$lab103.lab103c2",
			"customer":  "$lab14.lab14c3",
			"createdAt": "$createdAt",

			"status": bson.M{
				"$cond": bson.A{
					bson.M{
						"$and": bson.A{
							bson.M{
								"$gt": bson.A{
									bson.M{"$size": "$tests"},
									0,
								},
							},
							bson.M{
								"$eq": bson.A{
									bson.M{
										"$size": bson.M{
											"$filter": bson.M{
												"input": "$tests",
												"as":    "test",
												"cond": bson.M{
													"$not": bson.M{
														"$in": bson.A{
															"$$test.lab57c8",
															bson.A{4, 6, "4", "6"},
														},
													},
												},
											},
										},
									},
									0,
								},
							},
						},
					},
					6,
					0,
				},
			},

			"tests": bson.M{
				"$map": bson.M{
					"input": "$tests",
					"as":    "test",
					"in": bson.M{
						"area": bson.M{
							"$substrCP": bson.A{
								bson.M{
									"$replaceAll": bson.M{
										"input": bson.M{
											"$replaceAll": bson.M{
												"input": bson.M{
													"$replaceAll": bson.M{
														"input": bson.M{
															"$replaceAll": bson.M{
																"input": bson.M{
																	"$replaceAll": bson.M{
																		"input": bson.M{
																			"$toUpper": "$$test.lab39.lab43.lab43c4",
																		},
																		"find":        "Á",
																		"replacement": "A",
																	},
																},
																"find":        "É",
																"replacement": "E",
															},
														},
														"find":        "Í",
														"replacement": "I",
													},
												},
												"find":        "Ó",
												"replacement": "O",
											},
										},
										"find":        "Ú",
										"replacement": "U",
									},
								},
								0,
								3,
							},
						},
						"testName":      "$$test.lab39.lab39c4",
						"testID":        "$$test.lab39.lab39c2",
						"created":       "$$test.createdAt",
						"validatedDate": "$$test.lab57c18",
						"testStatus":    "$$test.lab57c8",
					},
				},
			},
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
