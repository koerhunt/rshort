package main

import (
	"crypto/tls"
	"github.com/gin-gonic/gin"
	_ "github.com/heroku/x/hmetrics/onload"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"log"
	"net"
	"net/http"
	"os"
)

//MODELS

type EntryUrl struct {
	Url string
	Key string
	ExpireAt string
}


//ENDOMDELS
func makeMgoSession() (*mgo.Session, error){

	tlsConfig := &tls.Config{

	}


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

		//parse params
		type EntryUrl struct {
			Url     string `json:"url" binding:"required"`
			Key string `json:"key" binding:"required"`
		}

		var data EntryUrl
		c.Bind(&data)

		//set connection
		s, err := makeMgoSession()
		if err != nil {
			log.Fatal(err)
		}
		collection := s.DB("rshort").C("urls")


		result := EntryUrl{}
		err2 := collection.Find(bson.M{"key": data.Key}).One(&result)


		if err2 != nil {

			err = collection.Insert(&EntryUrl{Key: data.Key, Url: data.Url})
			if err != nil {
				log.Fatal(err)
				c.JSON(http.StatusUnprocessableEntity,"{message: 'Error saving record'}")

			}else{
				c.JSON(http.StatusNoContent,"{}")
			}

		}else{

			c.JSON(http.StatusUnprocessableEntity,"{message: 'Key is not available'}")
		}



	})

	router.Run(":" + port)
}