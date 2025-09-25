package roles_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/DaanV2/mechanus/server/engine/users/roles"
)

var _ = Describe("Role", func() {
	DescribeTable("given roles should have or inherit the role",
		func(grant, check roles.Role) {
			Expect(grant.Inherits(check)).To(BeTrue(), "%s should inherit %s", grant, check)
		},
		Entry("Admin should have viewer", roles.Admin, roles.Viewer),
		Entry("Admin should have user", roles.Admin, roles.User),
		Entry("Admin should have operator", roles.Admin, roles.Operator),
		Entry("Admin should have admin", roles.Admin, roles.Admin),
		Entry("Operator should have viewer", roles.Operator, roles.Viewer),
		Entry("Operator should have user", roles.Operator, roles.User),
		Entry("Operator should have operator", roles.Operator, roles.Operator),
		Entry("User should have viewer", roles.User, roles.Viewer),
		Entry("User should have user", roles.User, roles.User),
		Entry("Viewer should have viewer", roles.Viewer, roles.Viewer),
		Entry("Viewer should have viewer", roles.Viewer, roles.Device),
	)

	DescribeTable("given roles should not have or inherit the role",
		func(grant, check roles.Role) {
			Expect(grant.Inherits(check)).To(BeFalse(), "%s should inherit %s", grant, check)
		},
		Entry("Operator should not have admin", roles.Operator, roles.Admin),
		Entry("User should not have admin", roles.User, roles.Admin),
		Entry("User should not have operator", roles.User, roles.Operator),
		Entry("Viewer should not have admin", roles.Viewer, roles.Admin),
		Entry("Viewer should not have operator", roles.Viewer, roles.Operator),
		Entry("Viewer should not have user", roles.Viewer, roles.User),
	)
})
