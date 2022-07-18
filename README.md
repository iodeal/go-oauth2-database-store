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
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	dbstore "github.com/iodeal/go-oauth2-database-store/v4"
)

func main() {
	manager := manage.NewDefaultManager()

	// use mysql token store
	store := mysql.NewDefaultStore(
		mysql.NewConfig("postgres","host=localhost port=5432 user=postgres password=123456 dbname=postgres sslmode=disable"),
	)

	defer store.Close()

	manager.MapTokenStorage(store)
	// ...
}

```

## MIT License

```
Copyright (c) 2018 Lyric
Copyright (c) 2022 iodeal
```

[license-url]: http://opensource.org/licenses/MIT
[license-image]: https://img.shields.io/npm/l/express.svg


