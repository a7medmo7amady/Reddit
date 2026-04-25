package main 

import (
	"fmt"
	"log"
	"os"
	"time"
	"context"
	"github.com/IBM/sarama"
)


type Event struct{
	eventID string 
	eventType string
	service string
	timeStamp int64 

}
