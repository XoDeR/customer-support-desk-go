package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/XoDeR/customer-support-desk-go/internal/adapter/repository/postgres"
	"github.com/XoDeR/customer-support-desk-go/internal/domain/entity"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/config"
	"github.com/XoDeR/customer-support-desk-go/internal/infrastructure/database"
	"github.com/XoDeR/customer-support-desk-go/internal/usecase"
	"github.com/XoDeR/customer-support-desk-go/pkg/migration"
	"github.com/XoDeR/customer-support-desk-go/pkg/uuidv7"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := database.NewPostgresConnection(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := migration.NewManager(db, "migrations/app").Migrate(ctx); err != nil {
		log.Fatal(err)
	}

	repo := postgres.New(db)
	ensureUser := func(email, password string, role entity.Role) entity.User {
		if u, e := repo.GetUserByEmail(ctx, email); e == nil {
			return u
		}
		hash, e := usecase.HashPassword(password)
		if e != nil {
			log.Fatal(e)
		}
		u := entity.User{ID: uuidv7.New(), Email: email, PasswordHash: hash, Role: role, Status: "active"}
		if e := repo.CreateUser(ctx, u); e != nil {
			log.Fatal(e)
		}
		return u
	}

	admin := ensureUser(cfg.Admin.Email, cfg.Admin.Password, entity.RoleAdmin)
	agent := ensureUser(cfg.Agent.Email, cfg.Agent.Password, entity.RoleAgent)
	customer := ensureUser("customer@example.com", "customer-password-change-me", entity.RoleCustomer)

	teams, err := repo.Teams(ctx)
	if err != nil {
		log.Fatal(err)
	}
	var team entity.Team
	if len(teams) == 0 {
		team, err = repo.CreateTeam(ctx, "General Support")
		if err != nil {
			log.Fatal(err)
		}
	} else {
		team = teams[0]
	}

	tags, _ := repo.Tags(ctx)
	if len(tags) == 0 {
		for _, name := range []string{"billing", "login", "urgent-followup"} {
			if _, e := repo.CreateTag(ctx, name); e != nil && !strings.Contains(e.Error(), "duplicate") {
				log.Printf("tag %s: %v", name, e)
			}
		}
	}

	replies, _ := repo.CannedReplies(ctx)
	if len(replies) == 0 {
		_, _ = repo.CreateCannedReply(ctx, entity.CannedReply{
			ID:        uuidv7.New(),
			Title:     "Thanks for contacting us",
			Body:      "Thanks for reaching out. We're looking into this and will update you shortly.",
			CreatedBy: agent.ID,
		})
	}

	tickets, err := repo.ListTickets(ctx, customer, 5, 0, "")
	if err != nil {
		log.Fatal(err)
	}
	if len(tickets) == 0 {
		due := time.Now().UTC().Add(24 * time.Hour)
		t := entity.Ticket{
			ID:          uuidv7.New(),
			Title:       "Cannot reset password",
			Description: "The reset email never arrives. Seeded demo ticket.",
			CustomerID:  customer.ID,
			AssigneeID:  &agent.ID,
			TeamID:      &team.ID,
			Status:      entity.StatusOpen,
			Priority:    entity.PriorityHigh,
			Category:    entity.CategoryAccount,
			SLADueAt:    &due,
		}
		if e := repo.CreateTicket(ctx, t); e != nil {
			log.Fatal(e)
		}
		_ = repo.Audit(ctx, t.ID, admin.ID, "ticket.created", t)
		_ = repo.AddComment(ctx, entity.Comment{
			ID:         uuidv7.New(),
			TicketID:   t.ID,
			AuthorID:   customer.ID,
			Body:       "I tried twice yesterday with no email.",
			Visibility: "public",
		})
	}

	fmt.Println("Seed complete")
	fmt.Printf("  admin:    %s\n", admin.Email)
	fmt.Printf("  agent:    %s\n", agent.Email)
	fmt.Printf("  customer: customer@example.com / customer-password-change-me\n")
	fmt.Printf("  team:     %s\n", team.Name)
}
