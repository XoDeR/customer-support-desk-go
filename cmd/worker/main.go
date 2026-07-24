package main

import (
	"context"
	"encoding/json"
	"github.com/XoDeR/customer-support-desk-go/internal/adapter/ws"
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/config"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/database"
	redisinfrastructure "github.com/XoDeR/customer-support-desk-go/internal/infrastructure/redis"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
	"log"
	"time"
)

func main() {
	c, e := config.Load()
	if e != nil {
		log.Fatal(e)
	}
	db, e := database.NewPostgresConnection(c)
	if e != nil {
		log.Fatal(e)
	}
	defer db.Close()
	redisClient, e := redisinfrastructure.New(c)
	if e != nil {
		log.Fatal(e)
	}
	defer redisClient.Close()
	publisher := ws.NewPublisher(redisClient)
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			rows, err := db.QueryContext(context.Background(), `UPDATE tickets SET breached_at=now() WHERE breached_at IS NULL AND sla_due_at < now() AND sla_paused_at IS NULL AND status NOT IN ('resolved','closed') RETURNING id`)
			if err != nil {
				log.Printf("sla breach job: %v", err)
				continue
			}
			for rows.Next() {
				var id uuidv7.UUID
				if err := rows.Scan(&id); err != nil {
					continue
				}
				_, _ = db.ExecContext(context.Background(), `INSERT INTO audit_events(id,ticket_id,event_type,payload) VALUES($1,$2,'sla.breached','{}')`, uuidv7.New(), id)
				_ = publisher.Publish(context.Background(), entity.DomainEvent{Type: "sla.breached", TicketID: id})
			}
			rows.Close()
		default:
			job, err := redisClient.BRPop(context.Background(), time.Second, "support.escalations").Result()
			if err != nil {
				continue
			}
			var event entity.DomainEvent
			if err := json.Unmarshal([]byte(job[1]), &event); err != nil {
				log.Printf("invalid escalation job: %v", err)
				continue
			}
			_ = publisher.Publish(context.Background(), event)
		}
	}
}
