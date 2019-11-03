package main

import (
	"integration_framework/application"
	_ "integration_framework/plugins/docker_compose"
	_ "integration_framework/plugins/graphql"
	_ "integration_framework/plugins/http_server"
	_ "integration_framework/plugins/postgres"
	"log"
	"math/rand"
	"time"
)

// TODO хорошо бы вынести нейминг сервисов и их файлов из самих сервисов в docker_compose плагин чтобы он диктовал условия, а не плагины

func main() {
	rand.Seed(time.Now().UnixNano())

	app, err := application.New()
	if err != nil {
		log.Fatal(err)
	}
	err = app.Start()
	if err != nil {
		log.Fatal(err)
	}
}

/*
==== #"api_v1 mutation confirm_payment should return error for nonexisting payment"
unable to check request: invalid response: invalid value of key ".errors.[0].message".
 Actual value  : "Auth token invalid" (string)
 Expected value: "Payment not found" (string)
==== #"api_v1 mutation confirm_payment should return error for payment with status cart_assigned but without assigned cart"
application_1          | 2019-11-02 16:04:23 [GIN] 172.18.0.1 POST /api -- 200 (7.992ms) headers map[Accept-Encoding:[gzip] Auth-Token:[test-auth-token] Content-Length:[84] User-Agent:[Go-http-client/1.1]]
unable to check request: invalid response: unexpected value of key ".errors.[0].path": <nil>
==== #"api_v1 mutation confirm_payment should return error and delete payment from db for expired payment"
unable to check request: invalid response: invalid value of key ".errors.[0].extensions.code".
 Actual value  : "auth_error:auth_token_invalid" (string)
 Expected value: "not_found_error:payment" (string)
==== #"api_v1 mutation confirm_payment should confirm payment"
unable to check request: invalid response: invalid value of key ".data".
 Actual value  : <nil> (<nil>)
 Expected value: map[string]interface {}{"confirm_payment":true} (map[string]interface {})
==== #"api_v1 mutation confirm_payment should return error for account with no selected masterpass card"
unable to check request: invalid response: invalid value of key ".errors.[0].message".
 Actual value  : "Auth token invalid" (string)
 Expected value: "Account have no selected masterpass card" (string)
==== #"api_v1 mutation confirm_payment should return error for account with invalid selected masterpass card uid"
unable to check request: invalid response: invalid value of key ".errors.[0].message".
 Actual value  : "Auth token invalid" (string)
 Expected value: "Card not found in masterpass account" (string)
==== #"api_v1 mutation confirm_payment should return error if account token invalid"
unable to check request: invalid response: invalid value of key ".errors.[0].message".
 Actual value  : "Auth token invalid" (string)
 Expected value: "Account not found" (string)
==== #"api_v1 mutation confirm_payment should return error for payment assigned to another account"
unable to check request: invalid response: unexpected value of key ".errors.[0].path": <nil>
==== #"api_v1 mutation confirm_payment should return error for payment with status other than cart_assigned"
unable to check request: invalid response: invalid value of key ".errors.[0].message".
 Actual value  : "Auth token invalid" (string)
 Expected value: "Invalid payment status" (string)
*/
