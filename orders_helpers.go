package main

import (
	"go.mongodb.org/mongo-driver/bson"
)

func getOrderStatusExpression() bson.M {
	return bson.M{
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
	}
}

func getRandomOpTimeExpression() bson.M {
	return bson.M{
		"$floor": bson.M{
			"$multiply": bson.A{
				bson.M{"$rand": bson.M{}},
				121,
			},
		},
	}
}

func getTestsExpression() bson.M {
	return bson.M{
		"$map": bson.M{
			"input": "$tests",
			"as":    "test",
			"in": bson.M{
				"area":          getAreaExpression(),
				"testName":      "$$test.lab39.lab39c4",
				"testID":        "$$test.lab39.lab39c2",
				"created":       "$$test.createdAt",
				"validatedDate": "$$test.lab57c18",
				"testStatus":    "$$test.lab57c8",
			},
		},
	}
}

func getAreaExpression() bson.M {
	return bson.M{
		"$substrCP": bson.A{
			getNormalizedAreaExpression(),
			0,
			3,
		},
	}
}

func getNormalizedAreaExpression() bson.M {
	return bson.M{
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
	}
}
