
# Create .env file en the root of the project

```
PORT=8080

MONGO_R0="xyz.mongodb.net:27017"
MONGO_R1="xyz.mongodb.net:27017"
MONGO_R2="xyz.mongodb.net:27017"
MONGO_DB="admin"
MONGO_USER="password"
MONGO_PWD="password"
```

#if you need generate grpc serivce

run:
```protoc -I grpc grpc/rshort.proto --go_out=plugins=grpc:grpc```