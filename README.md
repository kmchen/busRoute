# Bus Route Challenge

## Build && Run

```shell
$> go build main.go
$> ./main -busRoutePath={route_file}
or
$> ./main -shasumPath={path_to_shasum} -busRoutePath={route_file}
```

## Test

### Server

```shell
$> ./main -busRoutePath=sample_data
```

#### Route found 

```shell
$> curl -X GET \
  -H "Cache-Control: no-cache" \
    "http://localhost:8088/direct?dep_sid=270639&arr_sid=892937"
```
#### Route NOT found 

```shell
$> curl -X GET \
  -H "Cache-Control: no-cache" \
    "http://localhost:8088/direct?dep_sid=270639&arr_sid=892"
```
#### Update 

```shell
$> curl -v -X GET \
  -H "Cache-Control: no-cache" \
  "http://localhost:8088/update"
```
## Generate maximum sample routes
```shell
$> go run data.go
```
