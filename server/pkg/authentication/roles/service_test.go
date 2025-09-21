package roles_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/pkg/authentication/roles"
)

type mockRoleContainer struct {
	roles []string
}

// SetRoles implements roles.RoleContainer.
func (m *mockRoleContainer) SetRoles(newroles ...string) {
	m.roles = newroles
}

func (m *mockRoleContainer) GetRoles() []string {
	return m.roles
}

var _ = Describe("Service", func() {
	service := &roles.RoleService{}

	Context("ParseRole", func() {
		It("should parse valid roles", func() {
			validRoles := []string{"admin", "operator", "user", "viewer"}
			for _, role := range validRoles {
				parsedRole, err := roles.ParseRole(role)
				Expect(err).ToNot(HaveOccurred())
				Expect(parsedRole.String()).To(Equal(role))
			}
		})

		It("should return an error for unknown roles", func() {
			_, err := roles.ParseRole("unknown")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("unknown role: unknown"))
		})
	})

	Context("GrantRole", func() {

		It("should grant a role if not already granted", func() {
			grants := &mockRoleContainer{}
			service.GrantRole(grants, roles.Admin)
			Expect(grants.GetRoles()).To(ContainElement(roles.Admin.String()))
		})

		It("should not grant a role if already granted", func() {
			grants := &mockRoleContainer{roles: []string{roles.User.String()}}
			service.GrantRole(grants, roles.User)
			Expect(grants.GetRoles()).To(HaveLen(1))
		})
	})

	Context("HasRole", func() {
		DescribeTable("should check if a role is granted or inherited",
			func(grantedRoles []string, roleToCheck roles.Role, expected bool) {
				grants := &mockRoleContainer{roles: grantedRoles}
				hasRole := service.HasRole(grants, roleToCheck)
				Expect(hasRole).To(Equal(expected))
			},
			Entry("Admin has all roles", []string{roles.Admin.String()}, roles.Viewer, true),
			Entry("Operator has viewer", []string{roles.Operator.String()}, roles.Viewer, true),
			Entry("User has viewer", []string{roles.User.String()}, roles.Viewer, true),
			Entry("Viewer has viewer", []string{roles.Viewer.String()}, roles.Viewer, true),
		)
	})
})
