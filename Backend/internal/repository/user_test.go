package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/oms-saas/oms-saas-go/internal/apperror"
	"github.com/oms-saas/oms-saas-go/internal/testutil"
)

func TestUserManagementAndLogin(t *testing.T) {
	pool := testutil.SetupTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	suffix := strings.ReplaceAll(t.Name(), "/", "_")
	owner, tenant, err := repo.RegisterTenant(ctx, "User Test "+suffix, suffix+"@owner.local", "password123", "Owner")
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	staff, err := repo.CreateUser(ctx, tenant.ID, CreateUserInput{
		Email: suffix + "@staff.local", Password: "password123", FullName: "Staff User", Role: "staff",
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	found, err := repo.GetUserByEmailAndTenant(ctx, suffix+"@staff.local", tenant.Slug)
	if err != nil {
		t.Fatalf("login by slug: %v", err)
	}
	if found.ID != staff.ID {
		t.Fatalf("expected staff id %v, got %v", staff.ID, found.ID)
	}

	if err := repo.UpdateUserPassword(ctx, tenant.ID, owner.ID, "password123", "newpassword99"); err != nil {
		t.Fatalf("update password: %v", err)
	}

	users, err := repo.ListUsers(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}
	if len(users) < 2 {
		t.Fatalf("expected at least 2 users, got %d", len(users))
	}

	if err := repo.DeleteUser(ctx, tenant.ID, staff.ID); err != nil {
		t.Fatalf("delete user: %v", err)
	}

	err = repo.DeleteUser(ctx, tenant.ID, owner.ID)
	if err == nil {
		t.Fatal("expected error deleting owner")
	}
	if !strings.Contains(err.Error(), "cannot delete owner") {
		t.Fatalf("expected cannot delete owner, got %v", err)
	}
	_ = apperror.ErrValidation
}
