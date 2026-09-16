package services_test

import (
	"context"
	"errors"
	"testing"

	"folio/folio-core/domain"
	"folio/folio-core/ports"
	"folio/folio-core/services"
)

type workFixture struct {
	projects *fakeProjects
	plans    *fakePlans
	todos    *fakeTodos
	tickets  *fakeTickets
	planSvc  *services.PlanService
	todoSvc  *services.TodoService
	owner    ports.Actor
	viewer   ports.Actor
	outside  ports.Actor
	project  domain.ProjectID
}

func newWorkFixture(t *testing.T) *workFixture {
	t.Helper()
	projects := newFakeProjects()
	projects.items["p001"] = domain.Project{ID: "p001", Slug: "api", Name: "API", Members: []domain.Member{
		{UserID: "u-owner", Role: domain.RoleOwner},
		{UserID: "u-viewer", Role: domain.RoleViewer},
	}}
	// A second project the actor is not a member of, to catch leaks.
	projects.items["p002"] = domain.Project{ID: "p002", Slug: "other", Members: []domain.Member{
		{UserID: "u-someone", Role: domain.RoleOwner},
	}}

	plans := newFakePlans()
	todos := newFakeTodos()
	tickets := newFakeTickets()
	guard := services.NewProjectGuard(projects)
	clock := &fakeClock{now: testNow}

	todoSvc := services.NewTodoService(todos, plans, tickets, guard, clock, &seqIDs{prefix: "t"}, nopLogger{})
	planSvc := services.NewPlanService(plans, todos, tickets, todoSvc, guard, clock, &seqIDs{prefix: "pl"}, nopLogger{})

	return &workFixture{
		projects: projects, plans: plans, todos: todos, tickets: tickets,
		planSvc: planSvc, todoSvc: todoSvc, project: "p001",
		owner:   ports.Actor{UserID: "u-owner"},
		viewer:  ports.Actor{UserID: "u-viewer"},
		outside: ports.Actor{UserID: "u-stranger"},
	}
}

func TestCreatePlanWithTodosInOneCall(t *testing.T) {
	f := newWorkFixture(t)

	plan, err := f.planSvc.CreatePlan(context.Background(), f.owner, ports.CreatePlanInput{
		ProjectID: f.project,
		Title:     "Ship search",
		Goal:      "full text search over logs and docs",
		Todos: []ports.CreateTodoInput{
			{Title: "design the index"},
			{Title: "write the adapter"},
		},
	})
	if err != nil {
		t.Fatalf("create plan: %v", err)
	}

	if plan.Status != domain.PlanDraft {
		t.Fatalf("a plan should default to draft, got %q", plan.Status)
	}
	if plan.Progress.Total != 2 || plan.Progress.Done != 0 {
		t.Fatalf("want 0/2 progress, got %d/%d", plan.Progress.Done, plan.Progress.Total)
	}

	stored, err := f.todos.ListByPlan(context.Background(), plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) != 2 {
		t.Fatalf("want 2 todos persisted, got %d", len(stored))
	}
	if stored[0].Position == stored[1].Position {
		t.Fatal("todos created together must get distinct positions")
	}
	for _, td := range stored {
		if td.ProjectID != f.project {
			t.Fatalf("nested todo must inherit the plan's project, got %q", td.ProjectID)
		}
		if td.Status != domain.TodoPending || td.Priority != domain.PriorityMedium {
			t.Fatalf("want sane defaults, got %q/%q", td.Status, td.Priority)
		}
	}
}

func TestCreatePlanRequiresWriteAccess(t *testing.T) {
	f := newWorkFixture(t)

	_, err := f.planSvc.CreatePlan(context.Background(), f.viewer, ports.CreatePlanInput{ProjectID: f.project, Title: "No"})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not create plans, got %v", err)
	}

	_, err = f.planSvc.CreatePlan(context.Background(), f.outside, ports.CreatePlanInput{ProjectID: "p002", Title: "No"})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-member must get not-found, got %v", err)
	}
}

func TestGetPlanCarriesProgress(t *testing.T) {
	f := newWorkFixture(t)
	plan, err := f.planSvc.CreatePlan(context.Background(), f.owner, ports.CreatePlanInput{
		ProjectID: f.project, Title: "Ship",
		Todos: []ports.CreateTodoInput{{Title: "a"}, {Title: "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	first, err := f.todos.ListByPlan(context.Background(), plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.todoSvc.SetTodoStatus(context.Background(), f.owner, first[0].ID, domain.TodoDone); err != nil {
		t.Fatal(err)
	}

	got, err := f.planSvc.GetPlan(context.Background(), f.owner, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Progress.Done != 1 || got.Progress.Total != 2 {
		t.Fatalf("want 1/2, got %d/%d", got.Progress.Done, got.Progress.Total)
	}
}

func TestUpdatePlanValidatesStatus(t *testing.T) {
	f := newWorkFixture(t)
	plan, _ := f.planSvc.CreatePlan(context.Background(), f.owner, ports.CreatePlanInput{ProjectID: f.project, Title: "Ship"})

	bad := domain.PlanStatus("sideways")
	if _, err := f.planSvc.UpdatePlan(context.Background(), f.owner, plan.ID, ports.UpdatePlanInput{Status: &bad}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}

	good := domain.PlanActive
	got, err := f.planSvc.UpdatePlan(context.Background(), f.owner, plan.ID, ports.UpdatePlanInput{Status: &good})
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != domain.PlanActive {
		t.Fatalf("want active, got %q", got.Status)
	}
	if got.Title != "Ship" {
		t.Fatalf("omitted title must survive, got %q", got.Title)
	}
}

func TestDeletePlanDetachesItsTodos(t *testing.T) {
	f := newWorkFixture(t)
	plan, _ := f.planSvc.CreatePlan(context.Background(), f.owner, ports.CreatePlanInput{
		ProjectID: f.project, Title: "Ship", Todos: []ports.CreateTodoInput{{Title: "a"}},
	})

	if err := f.planSvc.DeletePlan(context.Background(), f.owner, plan.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	remaining, err := f.todoSvc.ListTodos(context.Background(), f.owner, f.project, domain.TodoFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 1 {
		t.Fatalf("deleting a plan must not destroy its todos, got %d", len(remaining))
	}
	if remaining[0].PlanID != "" {
		t.Fatalf("orphaned todo should be detached, still points at %q", remaining[0].PlanID)
	}
}

func TestCreateTodoRejectsPlanFromAnotherProject(t *testing.T) {
	f := newWorkFixture(t)
	f.plans.items["pl-foreign"] = domain.Plan{ID: "pl-foreign", ProjectID: "p002", Title: "Elsewhere"}

	_, err := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{
		ProjectID: f.project, PlanID: "pl-foreign", Title: "sneaky",
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("a todo must not join a plan in another project, got %v", err)
	}
}

func TestCreateTodosReportsPerItemFailures(t *testing.T) {
	f := newWorkFixture(t)

	res, err := f.todoSvc.CreateTodos(context.Background(), f.owner, f.project, []ports.CreateTodoInput{
		{Title: "good one"},
		{Title: "   "},
		{Title: "another good one"},
	})
	if err != nil {
		t.Fatalf("a batch with one bad item must not fail wholesale: %v", err)
	}
	if len(res.Created) != 2 {
		t.Fatalf("want 2 created, got %d", len(res.Created))
	}
	if len(res.Errors) != 1 || res.Errors[0].Index != 1 {
		t.Fatalf("want one error at index 1, got %+v", res.Errors)
	}
}

func TestListTodosDerivesBlocked(t *testing.T) {
	f := newWorkFixture(t)
	first, err := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{ProjectID: f.project, Title: "first"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{
		ProjectID: f.project, Title: "second", DependsOn: []domain.TodoID{first.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	got, err := f.todoSvc.ListTodos(context.Background(), f.owner, f.project, domain.TodoFilter{})
	if err != nil {
		t.Fatal(err)
	}
	byID := map[domain.TodoID]domain.Todo{}
	for _, td := range got {
		byID[td.ID] = td
	}
	if !byID[second.ID].Blocked {
		t.Fatal("second depends on an open first, so it must read as blocked")
	}

	if _, err := f.todoSvc.SetTodoStatus(context.Background(), f.owner, first.ID, domain.TodoDone); err != nil {
		t.Fatal(err)
	}
	got, _ = f.todoSvc.ListTodos(context.Background(), f.owner, f.project, domain.TodoFilter{})
	for _, td := range got {
		if td.ID == second.ID && td.Blocked {
			t.Fatal("finishing the dependency must clear blocked")
		}
	}
}

func TestGetTodoDerivesBlockedToo(t *testing.T) {
	f := newWorkFixture(t)
	first, _ := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{ProjectID: f.project, Title: "first"})
	second, _ := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{
		ProjectID: f.project, Title: "second", DependsOn: []domain.TodoID{first.ID},
	})

	got, err := f.todoSvc.GetTodo(context.Background(), f.owner, second.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Blocked {
		t.Fatal("a single fetched todo must carry the same derived blocked flag as the list")
	}
}

func TestSetTodoStatusRefusesToReopenCancelled(t *testing.T) {
	f := newWorkFixture(t)
	todo, _ := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{ProjectID: f.project, Title: "x"})

	if _, err := f.todoSvc.SetTodoStatus(context.Background(), f.owner, todo.ID, domain.TodoCancelled); err != nil {
		t.Fatal(err)
	}
	if _, err := f.todoSvc.SetTodoStatus(context.Background(), f.owner, todo.ID, domain.TodoPending); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error, got %v", err)
	}
}

func TestTodoWriteRequiresProjectAccess(t *testing.T) {
	f := newWorkFixture(t)
	todo, _ := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{ProjectID: f.project, Title: "x"})

	if _, err := f.todoSvc.SetTodoStatus(context.Background(), f.viewer, todo.ID, domain.TodoDone); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not write, got %v", err)
	}
	if _, err := f.todoSvc.GetTodo(context.Background(), f.outside, todo.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("non-member must get not-found, got %v", err)
	}
	if err := f.todoSvc.DeleteTodo(context.Background(), f.viewer, todo.ID); !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("viewer must not delete, got %v", err)
	}
}

func TestUpdateTodoParsesDueDate(t *testing.T) {
	f := newWorkFixture(t)
	todo, _ := f.todoSvc.CreateTodo(context.Background(), f.owner, ports.CreateTodoInput{ProjectID: f.project, Title: "x"})

	bad := "next tuesday"
	if _, err := f.todoSvc.UpdateTodo(context.Background(), f.owner, todo.ID, ports.UpdateTodoInput{DueDate: &bad}); !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("want validation error for an unparseable date, got %v", err)
	}

	good := "2026-12-01T09:00:00Z"
	got, err := f.todoSvc.UpdateTodo(context.Background(), f.owner, todo.ID, ports.UpdateTodoInput{DueDate: &good})
	if err != nil {
		t.Fatal(err)
	}
	if got.DueDate == nil || got.DueDate.Year() != 2026 {
		t.Fatalf("due date not stored: %+v", got.DueDate)
	}

	empty := ""
	got, err = f.todoSvc.UpdateTodo(context.Background(), f.owner, todo.ID, ports.UpdateTodoInput{DueDate: &empty})
	if err != nil {
		t.Fatal(err)
	}
	if got.DueDate != nil {
		t.Fatal("an empty due date must clear it")
	}
}

func TestTodoBlockedIsCorrectOnCreateAndList(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()

	blocker, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
		ProjectID: f.project, Title: "Decide the shape",
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Run("a todo created behind an open dependency reports blocked", func(t *testing.T) {
		got, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
			ProjectID: f.project, Title: "Build it",
			DependsOn: []domain.TodoID{blocker.ID},
		})
		if err != nil {
			t.Fatal(err)
		}
		if !got.Blocked {
			t.Error("want blocked on create")
		}
	})

	t.Run("a dependency excluded by the filter still blocks", func(t *testing.T) {
		if _, err := f.todoSvc.SetTodoStatus(ctx, f.owner, blocker.ID, domain.TodoInProgress); err != nil {
			t.Fatal(err)
		}
		todos, err := f.todoSvc.ListTodos(ctx, f.owner, f.project, domain.TodoFilter{
			Status: []domain.TodoStatus{domain.TodoPending},
		})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, todo := range todos {
			if todo.Title != "Build it" {
				continue
			}
			found = true
			if !todo.Blocked {
				t.Error("a dependency filtered out of the result must still block")
			}
		}
		if !found {
			t.Fatal("the dependent todo should be in a pending-only listing")
		}
	})
}

func TestBlockedIsDerivedWithoutRereadingTheNewTodo(t *testing.T) {
	f := newTicketFixture(t)
	ctx := context.Background()

	blocker, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
		ProjectID: f.project, Title: "Blocker",
	})
	if err != nil {
		t.Fatal(err)
	}

	created, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
		ProjectID: f.project, Title: "Dependent", DependsOn: []domain.TodoID{blocker.ID},
	})
	if err != nil {
		t.Fatal(err)
	}

	delete(f.todos.items, created.ID)

	again, err := f.todoSvc.CreateTodo(ctx, f.owner, ports.CreateTodoInput{
		ProjectID: f.project, Title: "Dependent again", DependsOn: []domain.TodoID{blocker.ID},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !again.Blocked {
		t.Error("blocked must be derived from the dependencies, not from finding the todo in a listing")
	}
}
