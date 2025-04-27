# scrutateur
Toolbox to deal with clients behavior modeling

## Installation 

To create go.mod, actions were: 
1. go get gorm.io/gorm
2. go get gorm.io/driver/postgres
3. go get -u github.com/gin-gonic/gin
4. go get -u github.com/golang-jwt/jwt/v5
5. go get github.com/google/uuid

## Endpoints

* **/status** just a string if up 
* **/login** expects a form with user and password, validates auth and returns the authorization set with the correct bearer