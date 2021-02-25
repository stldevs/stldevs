module github.com/jakecoffman/stldevs

go 1.16

require (
	github.com/dghubble/gologin/v2 v2.2.0
	github.com/gin-gonic/gin v1.6.3
	github.com/go-playground/validator/v10 v10.3.0 // indirect
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/google/go-github v17.0.0+incompatible
	github.com/google/go-github/v32 v32.1.0
	github.com/jackc/fake v0.0.0-20150926172116-812a484cc733 // indirect
	github.com/jackc/pgx v3.6.2+incompatible
	github.com/jackc/pgx/v4 v4.8.1
	github.com/jakecoffman/crud v0.0.0-20210223153833-c61f51c0d945 // indirect
	github.com/jmoiron/sqlx v1.2.0
	github.com/json-iterator/go v1.1.10 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20200909081042-eff7692f9009 // indirect
	google.golang.org/protobuf v1.25.0 // indirect
	gopkg.in/yaml.v2 v2.3.0 // indirect
)

replace (
	github.com/jakecoffman/crud => ../crud
)
