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
## filter options:
- Like    Operation = "like"
- Eq      Operation = "eq"   
- Ne      Operation = "ne" // not equal
- In      Operation = "in"
- NotIn   Operation = "not in"
- IsNull  Operation = "is null"
- NotNull Operation = "not null"