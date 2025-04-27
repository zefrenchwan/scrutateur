# scrutateur
Toolbox to deal with clients behavior modeling

## Installation 

To create go.mod, actions were: 
1. go get github.com/jackc/pgx/v5
2. go get github.com/jackc/pgx/v5/pgxpool
3. go get -u github.com/gin-gonic/gin
4. go get -u github.com/golang-jwt/jwt/v5
5. go get github.com/google/uuid

## Endpoints

* **/status** just a string if up 
* **/login** expects a form with login and password, validates auth and returns the authorization set with the correct bearer. Example is `curl -i -X POST -H 'Content-Type: application/json' -d '{"login":"popo","password":"caca"}' localhost:3000/login`