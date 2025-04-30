# Redis Documentation

## How to Run

1. Check if golang is installed in your local system
2. Check if docker is installed in local and working
3. Pull the docker image which is hosted in docker repository using this command:
   ```bash
   docker pull dhanushcrueiso/acronis-redis:v0.1.2
   ```

4. Then to run and bind the ports to 3001 run the below command:
   ```bash
   docker run -d -p 3001:3000 --name acronis-redis dhanushcrueiso/acronis-redis:v0.1.2
   ```

5. In another repo get the client library using:
   ```bash
   go get github.com/dhanushcrueiso/coding-test@v0.1.4
   ```

This is the Link to Access the Postman Docs: [Postman Documentation Link]

## Client API Documentation

Then follow the below code documentation to talk with the redis in docker:

### Create client
```go
cacheClient := cache.NewClient("http://localhost:3001")
```

### String Operations

#### Insert Key with TTL (String)
```go
err := cacheClient.Set("user123", "dhanush", time.Second*10)
if err != nil {
    fmt.Println("Error setting value:", err)
}
```

#### Get Key String
```go
data, err := cacheClient.Get("user123")
if err != nil {
    fmt.Println("Error getting value:", err)
}
```

#### Update Key String
```go
err = cacheClient.Update("user:profile:123", "reddy")
if err != nil {
    fmt.Println("Error updating value:", err)
}
```

#### Remove Key String
```go
err = cacheClient.Remove("user123")
if err != nil {
    fmt.Println("Error getting value:", err)
}
```

### TTL Operations

#### Get TTL
```go
time1, err = cacheClient.GetTTL("user123")
if err != nil {
    fmt.Println("Error getting TTL:", err)
}
```

#### Set TTL
```go
err = cacheClient.SetTTL("user123", time.Second*100)
if err != nil {
    fmt.Println("Error setting TTL:", err)
}
```

### List Operations

#### Create List With TTL
```go
err := cacheClient.CreateList("user10", time.Second*40)
if err != nil {
    fmt.Println("Error creating the list")
}
```

#### Add Element To List
```go
err = cacheClient.Push("user10", "testing")
if err != nil {
    fmt.Println("Error getting value: one", err)
}
```

#### Get List Data
```go
data, err := cacheClient.GetList("user10")
if err != nil {
    fmt.Println("Error getting value:", err)
}
```

#### POP Data From List
```go
test, err := cacheClient.Pop("user10")
if err != nil {
    fmt.Println("Error getting value:", err)
}
```

#### Remove List
```go
err = cacheClient.RemoveList("user10")
if err != nil {
    fmt.Println("Error deleting value:", err)
}
```