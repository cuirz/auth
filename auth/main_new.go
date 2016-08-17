package auth

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"
	"config"
	"db"
)

type Person struct {
	Name string `redis:"name"`
	Age  int    `redis:"age"`
}

func RetrievePerson(conn redis.Conn, id string) (Person, error) {
	var person Person

	values, err := redis.Values(conn.Do("HGETALL", fmt.Sprintf("person:%s", id)))
	if err != nil {
		return person, err
	}

	err = redis.ScanStruct(values, &person)
	return person, err
}

func main() {
	// Simulate command result
	//初始化配置
	//config.Init()
	db.GetRedis()
	redConn := db.Get()


	conn := redigomock.NewConn()
	cmd := conn.Command("HGETALL", "person:1").ExpectMap(map[string]string{
		"name": "Mr. Johson",
		"age":  "42",
	})

	person, err := RetrievePerson(redConn, "1")
	fmt.Println(person)
	if err != nil {
		fmt.Println(err)
		return
	}

	if conn.Stats(cmd) != 1 {
		fmt.Println("Command was not used first")
		return
	}

	if person.Name != "Mr. Johson" {
		fmt.Printf("Invalid name. Expected 'Mr. Johson' and got '%s'\n", person.Name)
		return
	}

	if person.Age != 42 {
		fmt.Printf("Invalid age. Expected '42' and got '%d'\n", person.Age)
		return
	}

	// Simulate command error

	conn.Clear()
	cmd = conn.Command("HGETALL", "person:1").ExpectError(fmt.Errorf("Simulate error!"))

	person, err = RetrievePerson(conn, "1")
	if err == nil {
		fmt.Println("Should return an error!")
		return
	}

	if conn.Stats(cmd) != 1 {
		fmt.Println("Command was not used")
		return
	}

	fmt.Println("Success!")
}