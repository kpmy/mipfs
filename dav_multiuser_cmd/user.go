package main

import (
	"github.com/ant0ine/go-json-rest/rest"
	"github.com/tv42/zbase32"
	"golang.org/x/crypto/bcrypt"
	"net/http"
)

type User struct {
	Login    string
	Password string
}

func regHandler() http.Handler {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)

	router, _ := rest.MakeRouter(rest.Post("/user", func(resp rest.ResponseWriter, req *rest.Request) {
		user := &User{}
		if err := req.DecodeJsonPayload(user); err == nil {
			if user.Login != "" && user.Password != "" {
				if _, err := KV.Read(user.Login); err != nil {
					pwd, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.MinCost)
					KV.Write(zbase32.EncodeToString([]byte(user.Login)), pwd)
				} else {
					rest.Error(resp, "wrong user", 400)
				}
			} else {
				rest.Error(resp, "empty data", 400)
			}
		} else {
			rest.Error(resp, "wrong data", 400)
		}
	}))

	api.SetApp(router)
	return api.MakeHandler()
}
