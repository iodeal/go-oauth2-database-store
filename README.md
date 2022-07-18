# Database storage for [OAuth 2.0](https://github.com/go-oauth2/oauth2)
Inherited from [MySQL storage for OAuth 2.0](https://github.com/go-oauth2/mysql)

[![License][license-image]][license-url]

## Install

``` bash
$ go get -v github.com/iodeal/go-oauth2-database-store/v4
```

## Usage

``` go
package main

import (
	"context"
	"fmt"
	"time"

	dbstore "github.com/iodeal/go-oauth2-database-store/v4"

	"github.com/go-oauth2/oauth2/v4/models"

	"github.com/go-oauth2/oauth2/v4/manage"

	_ "github.com/lib/pq"
)

func main() {
	manager := manage.NewDefaultManager()

	// use mysql token store
	store := dbstore.NewDefaultStore(
		dbstore.NewConfig("postgres", "host=localhost port=5432 user=postgres password=123456 dbname=postgres sslmode=disable"),
	)

	defer store.Close()

	manager.MapTokenStorage(store)
	info := &models.Token{
		ClientID:      "123",
		UserID:        "adb",
		RedirectURI:   "http://localhost/",
		Scope:         "all",
		Code:          "12_34_56",
		CodeCreateAt:  time.Now(),
		CodeExpiresIn: time.Second * 5,
	}
	err := store.Create(context.TODO(), info)
	if err != nil {
		log.Fatal("create store err:", err)
	} else {
		fmt.Println("create ok!")
	}
}

```

## MIT License

```
Copyright (c) 2018 Lyric
Copyright (c) 2022 iodeal
```

[license-url]: http://opensource.org/licenses/MIT
[license-image]: https://img.shields.io/npm/l/express.svg