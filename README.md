# db

A small collection of database utilities that make life easier for database use
and testing.

```go
import (
	"database/sql"

    "github.com/harrybrwn/db"
)

func main() {
    pool, err := sql.Open("postgres", "host=localhost password=testlab")
    if err != nil {
        panic(err)
    }
    db := db.New(pool, db.WithLogger(slog.Default()))
    // ...
}
```
