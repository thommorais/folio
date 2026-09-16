package rules_test

import (
	"errors"
	"strings"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/domain/rules"
)

func validTicket() domain.Ticket {
	return domain.Ticket{
		ProjectID: "p1",
		Slug:      "auth-token-expiry",
		Title:     "Tokens expire an hour early",
		Status:    domain.TicketOpen,
		Priority:  domain.PriorityMedium,
	}
}

func TestValidateTicket(t *testing.T) {
	t.Run("accepts a valid ticket", func(t *testing.T) {
		if err := rules.ValidateTicket(validTicket()); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("requires a project", func(t *testing.T) {
		ticket := validTicket()
		ticket.ProjectID = ""
		if err := rules.ValidateTicket(ticket); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("requires a title", func(t *testing.T) {
		ticket := validTicket()
		ticket.Title = "   "
		if err := rules.ValidateTicket(ticket); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("rejects a slug that is not kebab-case", func(t *testing.T) {
		for _, slug := range []string{"Has Space", "UPPER", "trailing-", ""} {
			ticket := validTicket()
			ticket.Slug = slug
			if err := rules.ValidateTicket(ticket); !errors.Is(err, domain.ErrValidation) {
				t.Errorf("slug %q: want validation error, got %v", slug, err)
			}
		}
	})

	t.Run("rejects an unknown status", func(t *testing.T) {
		ticket := validTicket()
		ticket.Status = "archived"
		if err := rules.ValidateTicket(ticket); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("rejects an unknown priority", func(t *testing.T) {
		ticket := validTicket()
		ticket.Priority = "urgent"
		if err := rules.ValidateTicket(ticket); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("rejects an over-long external ref", func(t *testing.T) {
		ticket := validTicket()
		ticket.ExternalRef = strings.Repeat("x", rules.RefMaxLen+1)
		if err := rules.ValidateTicket(ticket); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})
}

func TestCanTransitionTicket(t *testing.T) {
	t.Run("allows an ordinary move", func(t *testing.T) {
		if err := rules.CanTransitionTicket(domain.TicketOpen, domain.TicketInProgress); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("allows reopening a closed ticket", func(t *testing.T) {
		if err := rules.CanTransitionTicket(domain.TicketClosed, domain.TicketOpen); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("refuses to revive a cancelled ticket", func(t *testing.T) {
		if err := rules.CanTransitionTicket(domain.TicketCancelled, domain.TicketOpen); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("allows a no-op on a cancelled ticket", func(t *testing.T) {
		if err := rules.CanTransitionTicket(domain.TicketCancelled, domain.TicketCancelled); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("rejects an unknown target status", func(t *testing.T) {
		if err := rules.CanTransitionTicket(domain.TicketOpen, "archived"); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})
}

func TestSameProject(t *testing.T) {
	ticket := domain.Ticket{ID: "t1", ProjectID: "p1"}

	t.Run("accepts a child in the ticket's project", func(t *testing.T) {
		if err := rules.TicketBelongsTo(ticket, "p1"); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("rejects a child from another project", func(t *testing.T) {
		if err := rules.TicketBelongsTo(ticket, "p2"); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})
}

func TestApplyTicketBlocked(t *testing.T) {
	t.Run("blocks while a dependency is open", func(t *testing.T) {
		tickets := []domain.Ticket{
			{ID: "a", Status: domain.TicketOpen},
			{ID: "b", Status: domain.TicketOpen, DependsOn: []domain.TicketID{"a"}},
		}
		rules.ApplyTicketBlocked(tickets)
		if !tickets[1].Blocked {
			t.Fatal("b should be blocked by open a")
		}
		if tickets[0].Blocked {
			t.Fatal("a has no dependencies and must not be blocked")
		}
	})

	t.Run("clears once every dependency is terminal", func(t *testing.T) {
		tickets := []domain.Ticket{
			{ID: "a", Status: domain.TicketClosed},
			{ID: "b", Status: domain.TicketCancelled},
			{ID: "c", Status: domain.TicketOpen, DependsOn: []domain.TicketID{"a", "b"}},
		}
		rules.ApplyTicketBlocked(tickets)
		if tickets[2].Blocked {
			t.Fatal("c should not be blocked by terminal dependencies")
		}
	})

	t.Run("stays blocked while any one dependency is open", func(t *testing.T) {
		tickets := []domain.Ticket{
			{ID: "a", Status: domain.TicketClosed},
			{ID: "b", Status: domain.TicketOpen},
			{ID: "c", Status: domain.TicketOpen, DependsOn: []domain.TicketID{"a", "b"}},
		}
		rules.ApplyTicketBlocked(tickets)
		if !tickets[2].Blocked {
			t.Fatal("c should be blocked while b is open")
		}
	})

	t.Run("ignores a dependency that is not in the set", func(t *testing.T) {
		tickets := []domain.Ticket{{ID: "b", Status: domain.TicketOpen, DependsOn: []domain.TicketID{"ghost"}}}
		rules.ApplyTicketBlocked(tickets)
		if tickets[0].Blocked {
			t.Fatal("an unresolvable dependency must not block")
		}
	})
}

func TestCheckNoTicketCycle(t *testing.T) {
	graph := func(tickets ...domain.Ticket) []domain.Ticket { return tickets }

	t.Run("accepts a dependency that closes no loop", func(t *testing.T) {
		set := graph(
			domain.Ticket{ID: "a"},
			domain.Ticket{ID: "b", DependsOn: []domain.TicketID{"a"}},
		)
		if err := rules.CheckNoTicketCycle(set[1], set); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("rejects a ticket that depends on itself", func(t *testing.T) {
		candidate := domain.Ticket{ID: "a", DependsOn: []domain.TicketID{"a"}}
		if err := rules.CheckNoTicketCycle(candidate, graph(candidate)); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("rejects a two-ticket cycle", func(t *testing.T) {
		candidate := domain.Ticket{ID: "a", DependsOn: []domain.TicketID{"b"}}
		set := graph(candidate, domain.Ticket{ID: "b", DependsOn: []domain.TicketID{"a"}})
		if err := rules.CheckNoTicketCycle(candidate, set); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("rejects a three-ticket cycle", func(t *testing.T) {
		candidate := domain.Ticket{ID: "a", DependsOn: []domain.TicketID{"b"}}
		set := graph(
			candidate,
			domain.Ticket{ID: "b", DependsOn: []domain.TicketID{"c"}},
			domain.Ticket{ID: "c", DependsOn: []domain.TicketID{"a"}},
		)
		if err := rules.CheckNoTicketCycle(candidate, set); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("accepts a diamond, where a shared dependency is not a loop", func(t *testing.T) {
		candidate := domain.Ticket{ID: "d", DependsOn: []domain.TicketID{"b", "c"}}
		set := graph(
			domain.Ticket{ID: "a"},
			domain.Ticket{ID: "b", DependsOn: []domain.TicketID{"a"}},
			domain.Ticket{ID: "c", DependsOn: []domain.TicketID{"a"}},
			candidate,
		)
		if err := rules.CheckNoTicketCycle(candidate, set); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("ignores a dependency that is not in the set", func(t *testing.T) {
		candidate := domain.Ticket{ID: "a", DependsOn: []domain.TicketID{"ghost"}}
		if err := rules.CheckNoTicketCycle(candidate, graph(candidate)); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}

func TestCheckNoTicketAncestry(t *testing.T) {
	t.Run("accepts a parent that is not a descendant", func(t *testing.T) {
		candidate := domain.Ticket{ID: "b", ParentID: "a"}
		set := []domain.Ticket{{ID: "a"}, candidate}
		if err := rules.CheckNoTicketAncestry(candidate, set); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})

	t.Run("rejects a ticket parented to itself", func(t *testing.T) {
		candidate := domain.Ticket{ID: "a", ParentID: "a"}
		if err := rules.CheckNoTicketAncestry(candidate, []domain.Ticket{candidate}); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})

	t.Run("rejects a ticket parented to its own descendant", func(t *testing.T) {
		candidate := domain.Ticket{ID: "a", ParentID: "b"}
		set := []domain.Ticket{candidate, {ID: "b", ParentID: "c"}, {ID: "c", ParentID: "a"}}
		if err := rules.CheckNoTicketAncestry(candidate, set); !errors.Is(err, domain.ErrValidation) {
			t.Fatalf("want validation error, got %v", err)
		}
	})
}

func TestValidateTicketGraphFields(t *testing.T) {
	t.Run("rejects a ticket that depends on itself", func(t *testing.T) {
		bad := validTicket()
		bad.ID = "tk1"
		bad.DependsOn = []domain.TicketID{"tk1"}
		if !errors.Is(rules.ValidateTicket(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("rejects a ticket parented to itself", func(t *testing.T) {
		bad := validTicket()
		bad.ID = "tk1"
		bad.ParentID = "tk1"
		if !errors.Is(rules.ValidateTicket(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("rejects an unknown wayfinder type", func(t *testing.T) {
		bad := validTicket()
		bad.Wayfinder = "cartography"
		if !errors.Is(rules.ValidateTicket(bad), domain.ErrValidation) {
			t.Fatal("want validation error")
		}
	})

	t.Run("accepts every known wayfinder type", func(t *testing.T) {
		for _, w := range []domain.WayfinderType{
			domain.WayfinderMap, domain.WayfinderResearch, domain.WayfinderPrototype,
			domain.WayfinderGrilling, domain.WayfinderTask,
		} {
			ok := validTicket()
			ok.Wayfinder = w
			if err := rules.ValidateTicket(ok); err != nil {
				t.Fatalf("%s: want nil, got %v", w, err)
			}
		}
	})

	t.Run("accepts an empty wayfinder type", func(t *testing.T) {
		if err := rules.ValidateTicket(validTicket()); err != nil {
			t.Fatalf("want nil, got %v", err)
		}
	})
}
