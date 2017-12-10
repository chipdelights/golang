//Author : Pavani Boga ( pavanianapala@gmail.com )
//Date : 12/09/2017
//Purpose: Given the city, state name to this slack bot it will be querying yahoo weather api to fetch the weather info for that city and show it in slack channel
package main

import (
          "os"
          "fmt"
          "log"
          "encoding/json"
          "regexp"
          "net/http"
          "io/ioutil"
          "github.com/nlopes/slack"
       )

func getWeather(region string) string{
    url := `https://query.yahooapis.com/v1/public/yql?q=select%20*%20from%20weather.forecast%20where%20woeid%20in%20(select%20woeid%20from%20geo.places(1)%20where%20text%3D%22` + region + `%22)&format=json`
    response,err := http.Get(url)
    
    if err != nil {
        log.Fatal(err)
    }
   
    defer response.Body.Close()
    var data map[string]interface{}
    rbytes,_ := ioutil.ReadAll(response.Body)
    json.Unmarshal(rbytes,&data)
    title := fmt.Sprintf("%v",data["query"].(map[string]interface{})["results"].(map[string]interface{})["channel"].(map[string]interface{})["description"])
    desc := data["query"].(map[string]interface{})["results"].(map[string]interface{})["channel"].(map[string]interface{})["item"].(map[string]interface{})["forecast"].([]interface{})[0].(map[string]interface{})
    
    date,_ := desc["date"].(string)
    text,_ := desc["text"].(string)
    tempLow,_ := desc["low"].(string)
    tempHigh,_ := desc["high"].(string)
                 
    weather_data := fmt.Sprintf("%s\nDate: %s\nWeather: %v\nTemp Low: %v\t Temp High: %v\n", title, date, text, tempLow, tempHigh)
  
    return weather_data
}   

func main() {
    token := os.Getenv("SLACK_TOKEN")
    api := slack.New(token)
    logger := log.New(os.Stdout,"weather-bot: ",log.Lshortfile|log.LstdFlags)
    slack.SetLogger(logger)
    //api.SetDebug(true)   
   
    rtm := api.NewRTM()
    go rtm.ManageConnection()

    var botID string

Loop:
   for {
       select {
           case msg := <-rtm.IncomingEvents:
               switch event := msg.Data.(type) {
                   
                   case *slack.ConnectedEvent:
                       botID = event.Info.User.ID
 
                   case *slack.MessageEvent:
                       if event.User != "" {
                           r1 := regexp.MustCompile(`<@(.*?)>`)
                           if r1.FindStringSubmatch(event.Text)[1] == botID {
                               r2 := regexp.MustCompile(`(\S+)\s+(\S+)`)
                               place := r2.FindStringSubmatch(event.Text)[2]
                               response := getWeather(place)
                               params := slack.PostMessageParameters{}
                               api.PostMessage(event.Channel,response,params)
                           }
                      }
                   case *slack.RTMError:
                       fmt.Println("Error : %s\n", event.Error())

                  case *slack.InvalidAuthEvent:
                       fmt.Println("Invalid credentials")
                       break Loop

                  default:
             }
       }
   }
                    
}
