# scrutateur
Backend server to deal with patterns. 

## Installation 

### With docker compose 
start docker instances with compose: `docker compose -f 'compose.yaml' up -d --build`

### Test your application

If you want to test your installation: 
1. go to clients/
2. launch main: `go run main.go`

Expected result should look like 

```
Hello root (took  32.0662ms )
```

## Features

### Endpoints

* **/status** just a string if up 
* **/login** expects a form with login and password, validates auth and returns the authorization set with the correct bearer. Example is `curl -i -X POST -H 'Content-Type: application/json' -d '{"login":"root","password":"secret"}' localhost:3000/login`
* **/user/whoami/** displays user name if auth is valid and role allows it
* **/user/password** changes current user's password
* **/admin/user/create** will create an user (with no role)

### Security

This project is not intented to run on production as is. 
Code deals with basic security (just enough not to be ridiculous) and focuses more on features. 
Weak points are secrets protection and default user mechanism. 
**Adapt my code for your context, contact your administrator or security expert before pushing any of this code to production**

Security features so far:
* Role based mechanism
* user auth based on JWT, password is stored as a sha256 hash, no salt

### Client

There is a golang client to perform client calls 

## Architecture

1. endpoints are either unprotected (login and status) or protected (with an auth check mechanism and access to pages are based on roles)
2. Storage for auth is based on a relational database. 
3. Sessions are based on a cache, a session-id header is expected once user is connected

### Dependencies

To create go.mod, actions were: 
1. go get github.com/jackc/pgx/v5
2. go get github.com/jackc/pgx/v5/pgxpool
3. go get -u github.com/gin-gonic/gin
4. go get -u github.com/golang-jwt/jwt/v5
5. go get github.com/google/uuid
6. go get github.com/redis/go-redis/v9   

### The security model 

Users have a login and a password to prove their identity. 
Roles define what an user may do: grant or not, read or write. 
Hence, roles are: 
1. root: admin with ability to grant roles
2. admin: can perform admin functions such as creating users, changing roles, etc
3. editor: can see and edit non critical content 
4. reader: can see non critical content


Then, resources are grouped into functions by name. 
It depends on your configuration, but for instance: admin/create-user, admin/delete-user are basic admin operations forming a group of admin actions. 
Resources need some authorizations for users to connect to. 
For instance, admin/... expect admin or even root users. 


Users have roles too, on a group of resources. 