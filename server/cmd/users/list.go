package userscmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/DaanV2/mechanus/server/internal/components"
	"github.com/DaanV2/mechanus/server/pkg/models"
	"github.com/DaanV2/mechanus/server/pkg/terminal"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
)

// users/listCmd represents the users/list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the users currently registered",
	RunE:  ListUsers,
}

func init() {
	usersCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// users/listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// users/listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func ListUsers(cmd *cobra.Command, args []string) error {
	userService := components.UserService()
	userTable := terminal.NewTable(displayUser)

	userTable.SetColumns(
		table.Column{
			Title: "ID",
			Width: 10,
		},
		table.Column{
			Title: "Name",
			Width: 10,
		},
		table.Column{
			Title: "Roles",
			Width: 10,
		},
		table.Column{
			Title: "Campaigns",
			Width: 10,
		},
		table.Column{
			Title: "Created At",
			Width: 10,
		},
		table.Column{
			Title: "Updated At",
			Width: 10,
		},
		table.Column{
			Title: "Deleted At",
			Width: 10,
		},
	)

	var uerr error

	go func() {
		for id := range userService.Ids() {
			u, err := userService.Get(id)
			if err != nil {
				uerr = errors.Join(uerr, fmt.Errorf("error retrieving user %v: %w", id, err))
				continue
			}

			userTable.AddItem(&u)
			userTable.AutoWidth()
		}
	}()

	p := tea.NewProgram(userTable)
	if _, err := p.Run(); err != nil {
		return errors.Join(uerr, err)
	}

	return uerr
}

const (
	TIME_FORMAT = time.RFC850
)

func displayUser(user *models.User) []string {
	return []string{
		user.ID,
		user.Name,
		strings.Join(user.Roles, ","),
		strings.Join(user.Campaigns, ","),
		user.CreatedAt.Format(TIME_FORMAT),
		user.UpdatedAt.Format(TIME_FORMAT),
		timeString(user.DeletedAt),
	}
}

func timeString(t *time.Time) string {
	if t == nil {
		return "-"
	}

	return t.Format(TIME_FORMAT)
}
