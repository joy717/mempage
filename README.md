# Description
mempage is a simple pagination library for golang.
for client request data pagination.

# Feature
- Support filter 
- Support sort 
- Support page
```
{
  "filters": [
    {
      "key": "name",
      "op": "like",
      "values": [
        ""
      ]
    }
  ],
  "page": 1,
  "pageSize": 10,
  "sorts": [
    {
      "asceding": false,
      "key": "createTime"
    }
  ]
}
```

**Important**:

`key` (for filter and sort) is `the field name of json tag`. for example, we have a struct like this:

```
type User struct {
  Id   int64 `json:"id"`
  Name string `json:"my_name"`
  CreateTime time.Time `json:"createTime"`
}
```

if we want to sort by CreateTime, then the value of `key` is `createTime`.

if we want to filter by Name, then the value of `key` is `my_name`.


## Filter op(options):
- Like    Operation = "like"
- Eq      Operation = "eq"   
- Ne      Operation = "ne" // not equal
- In      Operation = "in"
- NotIn   Operation = "not in"
- IsNull  Operation = "is null"
- NotNull Operation = "not null"

## Example

```
package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/joy717/mempage"
)

type User struct {
	Id         int64     `json:"id"`
	Name       string    `json:"name"`
	CreateTime time.Time `json:"createTime"`
}

func main() {
	// filter data by name, which name contains "user"
	// sort by createTime desc.  
	// sorts[0].asceding = false means desc.
	pageReqStr := `
{
  "filters": [
    {
      "key": "name",
      "op": "like",
      "values": [
        "user"
      ]
    }
  ],
  "page": 1,
  "pageSize": 10,
  "sorts": [
    {
      "asceding": false,
      "key": "createTime"
    }
  ]
}
`

	// parse to page struct
	page := new(mempage.Page)
	if err := json.Unmarshal([]byte(pageReqStr), page); err != nil {
		panic(err)
	}

	// get data from somewhere
	u1 := User{
		Id:         1,
		Name:       "user1",
		CreateTime: time.Now(),
	}
	u2 := User{
		Id:         2,
		Name:       "user2",
		CreateTime: time.Now().Add(time.Hour),
	}
	u3 := User{
		Id:         3,
		Name:       "other",
		CreateTime: time.Now().Add(time.Hour * 2),
	}
	users := []User{u1, u2, u3}

	originalSLice, err := json.Marshal(users)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(originalSLice))
	// result: [{"id":1,"name":"user1","createTime":"2024-03-21T16:46:24.78192+08:00"},{"id":2,"name":"user2","createTime":"2024-03-21T17:46:24.78192+08:00"},{"id":3,"name":"other","createTime":"2024-03-21T18:46:24.78192+08:00"}]

	// pass data to mempage.Page to filter,sort,pagination
	page.FillResultAll(users)

	// the final data is in page.Result
	fmt.Println(page.Result)

	// print as json
	resultJsonBytes, err := json.Marshal(page.Result)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resultJsonBytes))
	// result: [{"id":2,"name":"user2","createTime":"2024-03-21T17:43:29.586374+08:00"},{"id":1,"name":"user1","createTime":"2024-03-21T16:43:29.586374+08:00"}]
}

```