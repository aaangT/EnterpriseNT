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

func getOrderSumStatusExpression(testsInput interface{}) bson.M {
	return bson.M{
		"$let": bson.M{
			"vars": bson.M{
				"statusList": bson.M{
					"$map": bson.M{
						"input": testsInput,
						"as":    "test",
						"in": bson.M{
							"$toString": "$$test.testStatus",
						},
					},
				},
			},
			"in": bson.M{
				"$switch": bson.M{
					"branches": bson.A{
						bson.M{
							"case": bson.M{
								"$eq": bson.A{
									bson.M{
										"$size": "$$statusList",
									},
									0,
								},
							},
							"then": 10,
						},
						bson.M{
							"case": getAllStatusCondition("0"),
							"then": 10,
						},
						bson.M{
							"case": getAllStatusCondition("1"),
							"then": 11,
						},
						bson.M{
							"case": getAllStatusCondition("2"),
							"then": 12,
						},
						bson.M{
							"case": getAllStatusCondition("5"),
							"then": 13,
						},
						bson.M{
							"case": getAllStatusInCondition(bson.A{"4", "6"}),
							"then": 14,
						},
						bson.M{
							"case": getAnyStatusInCondition(bson.A{"4", "6"}),
							"then": 15,
						},
						bson.M{
							"case": getAnyStatusInCondition(bson.A{"5"}),
							"then": 16,
						},
						bson.M{
							"case": getAnyStatusInCondition(bson.A{"2"}),
							"then": 17,
						},
						bson.M{
							"case": bson.M{
								"$and": bson.A{
									getAnyStatusInCondition(bson.A{"0"}),
									getAnyStatusInCondition(bson.A{"1"}),
								},
							},
							"then": 18,
						},
					},
					"default": 19,
				},
			},
		},
	}
}

func getStatusListHasDataCondition() bson.M {
	return bson.M{
		"$gt": bson.A{
			bson.M{
				"$size": "$$statusList",
			},
			0,
		},
	}
}

func getAllStatusCondition(status string) bson.M {
	return bson.M{
		"$and": bson.A{
			getStatusListHasDataCondition(),
			bson.M{
				"$eq": bson.A{
					bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$$statusList",
								"as":    "status",
								"cond": bson.M{
									"$ne": bson.A{
										"$$status",
										status,
									},
								},
							},
						},
					},
					0,
				},
			},
		},
	}
}

func getAllStatusInCondition(validStatuses bson.A) bson.M {
	return bson.M{
		"$and": bson.A{
			getStatusListHasDataCondition(),
			bson.M{
				"$eq": bson.A{
					bson.M{
						"$size": bson.M{
							"$filter": bson.M{
								"input": "$$statusList",
								"as":    "status",
								"cond": bson.M{
									"$not": bson.A{
										bson.M{
											"$in": bson.A{
												"$$status",
												validStatuses,
											},
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
	}
}

func getAnyStatusInCondition(validStatuses bson.A) bson.M {
	return bson.M{
		"$gt": bson.A{
			bson.M{
				"$size": bson.M{
					"$filter": bson.M{
						"input": "$$statusList",
						"as":    "status",
						"cond": bson.M{
							"$in": bson.A{
								"$$status",
								validStatuses,
							},
						},
					},
				},
			},
			0,
		},
	}
}
