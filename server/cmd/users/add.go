package userscmd

import (
	"errors"

	"github.com/DaanV2/mechanus/server/internal/components"
	"github.com/DaanV2/mechanus/server/pkg/extensions/ptr"
	xrand "github.com/DaanV2/mechanus/server/pkg/extensions/rand"
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/log"
	"github.com/spf13/cobra"
)

// users/listCmd represents the users/list command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "List all the users currently registered",
	RunE:  AddUser,
}

func init() {
	usersCmd.AddCommand(addCmd)
}

func AddUser(cmd *cobra.Command, args []string) error {
	var username *string = ptr.To("")
	var password *string = ptr.To(xrand.MustID(4))
	var roles *[]string = ptr.To([]string{"user"})

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("username").
				Value(username).
				CharLimit(32).
				Inline(true),
			huh.NewInput().
				Title("std password").
				Value(password).
				CharLimit(32).
				Inline(true),
			huh.NewMultiSelect[string]().
				Title("roles").
				Value(roles).
				Description("which roles the user should have").
				Options(huh.NewOptions("user", "admin")...),
		).Title("Create new user"),
	)

	err := form.Run()
	if err != nil {
		return err
	}
	if username == nil || password == nil || roles == nil {
		return errors.New("username, password or roles need to have been set")
	}

	userService := components.NewUserService()

	user := models.User{
		Name:         *username,
		Roles:        *roles,
		Campaigns:    []string{},
		PasswordHash: []byte(*password),
	}

	log.Info("creating user", "name", user.Name, "roles", user.Roles)
	u, err := userService.Create(user)

	if err != nil {
		log.Error("failed to create user", "error", err)
		return err
	}

	log.Info("created user",
		"id", u.ID,
		"name", u.Name,
		"roles", u.Roles,
		"campaigns", u.Campaigns,
	)
	return nil
}
