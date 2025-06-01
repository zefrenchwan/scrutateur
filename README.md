# scrutateur
Backend server to deal with patterns. 

## Installation 

### Configure options

If you add `REDIS_URL` as an environment variable, then you may use a redis cache in your code. 
Otherwise, code will only use a relational database (mandatory). 

### With docker compose 
start docker instances with compose: `docker compose -f 'compose.yaml' up -d --build`

### Test your application

If you want to test your installation: 
1. go to clients/
2. launch main: `go run main.go`

Expected result should look like 

```
Hello root (took  32.0662ms )
...
```


Any failure in that part means a feature at least is not available.

### Steps before pushing to production 

1. Change passwords (sql/) 
2. Change deployment properties (compose.yaml)
3. Change input validators (services.validators)
4. Audit code
5. Then, deliver 

## Features

### Endpoints

#### About (need no auth)

* **/app/static/changelog.txt** versions and news

#### Infrastructure (need no auth) 
* **/status** just a string if up

#### Unprotected operations 
* **/login** expects a form with login and password, validates auth and returns the authorization set with the correct bearer. Example is `curl -i -X POST -H 'Content-Type: application/json' -d '{"name":"root","password":"secret"}' localhost:3000/login`

#### Self group: actions from current user to current user 
* **/self/user/whoami/** displays user name if auth is valid and role allows it
* **/self/user/password** changes current user's password

#### Management operations on users

* **/manage/user/create** creates an user (with no role)
* **/manage/user/{username}/delete** deletes an user by name (no matter user's roles). Current user cannot delete current user
* **/manage/user/{username}/access/list** displays groups and matching roles for a given user
* **/manage/user/{username}/access/edit** changes groups and matching roles for a given user

### Security

This project is not intented to run on production as is. 
Code deals with basic security (roles, input validation, jwt) but was neither audited or approved by a security expert.  
Known weak points are secrets protection (no salt) and default user mechanism (root configuration by default). 
**Adapt my code for your context, contact your administrator or security expert before pushing any of this code to production**


### Client

There is a golang client to perform client calls. 
Available operations so far: 
* login 
* set password
* display current user name
* add user (needs admin role) and delete user (root only)
* display roles for user (admin) and change roles on groups (admin for admin, editor or reader, root for all roles)

## Architecture

1. endpoints are either unprotected (login and status) or protected (with an auth check mechanism and access to pages are based on roles)
2. Storage for auth is based on a relational database. 
3. Although not used, a redis cache is provided and may be set to active via a configuration mode

### Dependencies

To create go.mod, actions were: 
1. go get github.com/jackc/pgx/v5
2. go get github.com/jackc/pgx/v5/pgxpool
3. go get -u github.com/golang-jwt/jwt/v5
4. go get github.com/google/uuid
5. go get github.com/redis/go-redis/v9   

### The roles model 

Users have a login and a password to prove their identity. 
Roles define what an user may do: grant or not, read or write. 
Hence, roles are: 
1. root: admin with ability to grant critical roles and perform any operations (including critical ones) that makes sense
2. admin: can perform basic admin functions such as creating users, changing some roles, etc
3. editor: can see and edit non critical content 
4. reader: can see non critical content


Then, resources are grouped into functions by name. 
It depends on your configuration, but for instance: admin/create-user, admin/delete-user are basic admin operations forming a group of admin actions. 
Resources need some authorizations for users to connect to. 
For instance, admin/... expect admin or even root users. 


Users have roles too, on a group of resources. 


## FAQ 

### Wait, your `engines/` module looks a lot like gin gonic...

It does. 
I first wrote the code using Gin Gonic. 
And... 
I faced a bug / unexplained behavior: setting header in a non-final middleware did not work, header was just not set in the response.
So, I wrote a basic engine of my own. 

### What did you do to deal with security ? 

* Role based model
* input validators to prevent sql injection 

### I cloned your code for my project. I want to create a page, what are the main steps ?

1. Add your endpoint in `services` and link it to the `Init` function in services
2. Manage access into `03_content.sql` (the TODO part)
3. Add clients code in `clients/clients.go`

### Do you use generative AI to help you coding ? 

Nope, I am a dinosaur: I use only my brain, books, good ideas I read on the internet. 

### I have an idea for a feature, will you implement it ? 

Nope, I write open source code for that reason, with one of the most permissive licenses.
Feel free to clone the project and add your features. 