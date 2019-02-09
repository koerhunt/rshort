package main

import (
	"context"
	"crypto/tls"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	pb "github.com/koerhunt/rshort/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net"
	"net/http"
	"os"
	"github.com/joho/godotenv"
)

//MODELS

type EntryUrl struct {
	Url string
	Key string
	ExpireAt string
}

//parse params
type EntryUrlModel struct {
	Url     string `json:"url" binding:"required"`
	Key string `json:"key" binding:"required"`
}

const (
	grpcPort = ":50051"
)


type grpcServer struct{}

func (s *grpcServer) CutURL(ctx context.Context, in *pb.CutUrlRequest) (*pb.CutUrlReply, error) {
	log.Printf("Received: %v", in.Key)
	log.Printf("Received: %v", in.Url)

	data := EntryUrlModel{Url: in.Url, Key: in.Key}

	saved, err := CreateUrlEntry(&data)

	if saved {
		log.Printf("responded true")
		return &pb.CutUrlReply{Status: 200, Data: "https://rshort.herokuapp.com/url/"+data.Key}, nil
	}else{
		log.Printf("responded false")
		log.Print(err)
		return &pb.CutUrlReply{Status: 422, Data: "No se completo la accion"}, nil
	}
}

//ENDOMDELS
func makeMgoSession() (*mgo.Session, error){

	tlsConfig := &tls.Config{}

	dialInfo := &mgo.DialInfo{
		Addrs: []string{os.Getenv("MONGO_R0"),os.Getenv("MONGO_R1"),os.Getenv("MONGO_R2")},
		Database: os.Getenv("MONGO_DB"),
		Username: os.Getenv("MONGO_USER"),
		Password: os.Getenv("MONGO_PWD"),
		Source: "admin",
	}

	//Database
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}

	session, err := mgo.DialWithInfo(dialInfo)

	return session, err

}

func main() {


	loadEnv()

	go startGrpc()

	startWeb()

}

func startWeb()  {

	//LOAD ENV
	port := os.Getenv("PORT")

	if port == "" {
		log.Fatal("$PORT must be set")
	}

	//Setup Webserer
	router := gin.Default()
	router.Delims("<%=", "%>")
	router.Use(gin.Logger())

	//Static sources
	router.LoadHTMLGlob("templates/*.tmpl.html")
	router.Static("/static", "static")

	//Routes

	//GET /
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl.html", nil)
	})

	//ISSUSE entra en conflicto con la ruta /
	router.GET("/url/:shortkey", func(c *gin.Context) {

		s, err := makeMgoSession()
		if err != nil {
			log.Fatal(err)
		}else{
			log.Println("conected")
		}

		collection := s.DB("rshort").C("urls")

		key := c.Param("shortkey")
		result := EntryUrl{}
		err2 := collection.Find(bson.M{"key": key}).One(&result)

		if err2 != nil {
			c.Redirect(http.StatusSeeOther, "/not-found")
		}else{
			c.Redirect(http.StatusSeeOther, result.Url)
		}

	})

	//POST /save-key
	router.POST("/save-key", func(c *gin.Context){

		var data EntryUrlModel
		c.Bind(&data)

		saved, err := CreateUrlEntry(&data)

		if saved {
			c.JSON(http.StatusNoContent,"{}")
		}else{
			log.Fatal(err)
			c.JSON(http.StatusUnprocessableEntity,"{message: 'Error saving record'}")
		}


	})

	router.Run(":" + port)
}

func startGrpc()  {
	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("grpc failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterRshorterServer(s, &grpcServer{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func loadEnv(){
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func CreateUrlEntry(data *EntryUrlModel) (bool, error) {

	//set connection
	s, err := makeMgoSession()
	if err != nil {
		log.Fatal(err)
		return false, err
	}

	collection := s.DB("rshort").C("urls")

	result := EntryUrl{}

	//Check if exists
	err2 := collection.Find(bson.M{"key": data.Key}).One(&result)

	if err2 != nil {

		err = collection.Insert(&EntryUrl{Key: data.Key, Url: data.Url})
		if err != nil {
			log.Fatal(err)
			return false, err
		}else{
			return true, nil
		}

	}else{
		return false, err2
	}


}