package screens_test

import (
	"github.com/DaanV2/mechanus/server/mechanus/screens"
	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("ScreenID", func() {
	var (
		adminRole    = "admin"
		operatorRole = "operator"
		userRole     = "user"
		viewerRole   = "viewer"
		testUUID     = uuid.New().String()
		testID       screens.ScreenID
	)

	BeforeEach(func() {
		testUUID = uuid.New().String()
		testID = screens.NewScreenID(adminRole, testUUID)
	})

	Describe("Creation and Basic Properties", func() {
		It("should create a valid ScreenID with role and UUID", func() {
			for _, role := range []string{adminRole, operatorRole, userRole, viewerRole} {
				id := screens.NewScreenID(role, testUUID)
				Expect(id.String()).To(Equal(role + ":" + testUUID))
			}
		})

		It("should extract role correctly", func() {
			Expect(testID.Role()).To(Equal(adminRole))
		})

		It("should extract ID correctly", func() {
			Expect(testID.ID()).To(Equal(testUUID))
		})

		It("should panic with invalid format", func() {
			invalidID := screens.ScreenID("invalid-no-separator")
			Expect(func() { invalidID.Role() }).To(Panic())
			Expect(func() { invalidID.ID() }).To(Panic())
		})
	})

	Describe("Role Checks", func() {
		It("should correctly identify its role", func() {
			for _, role := range []string{adminRole, operatorRole, userRole, viewerRole} {
				id := screens.NewScreenID(role, testUUID)
				Expect(id.HasRole(role)).To(BeTrue())
				Expect(id.HasRole("invalid-role")).To(BeFalse())
			}
		})
	})

	Describe("ID Checks", func() {
		It("should correctly identify its UUID", func() {
			Expect(testID.HasID(testUUID)).To(BeTrue())
			Expect(testID.HasID(uuid.New().String())).To(BeFalse())
		})
	})

	Describe("Comparison Operations", func() {
		It("should correctly compare equal ScreenIDs", func() {
			id1 := screens.NewScreenID(adminRole, testUUID)
			id2 := screens.NewScreenID(adminRole, testUUID)
			Expect(id1.Equals(id2)).To(BeTrue())
		})

		It("should correctly compare different ScreenIDs", func() {
			id1 := screens.NewScreenID(adminRole, testUUID)
			id2 := screens.NewScreenID(adminRole, uuid.New().String())
			id3 := screens.NewScreenID(operatorRole, testUUID)
			
			Expect(id1.Equals(id2)).To(BeFalse(), "Different UUIDs should not be equal")
			Expect(id1.Equals(id3)).To(BeFalse(), "Different roles should not be equal")
		})
	})

	Describe("Empty Checks", func() {
		It("should correctly identify empty ScreenIDs", func() {
			emptyID := screens.ScreenID("")
			Expect(emptyID.IsEmpty()).To(BeTrue())
			Expect(testID.IsEmpty()).To(BeFalse())
		})
	})

	Describe("String Representation", func() {
		It("should return correct string format", func() {
			for _, role := range []string{adminRole, operatorRole, userRole, viewerRole} {
				id := screens.NewScreenID(role, testUUID)
				expected := role + ":" + testUUID
				Expect(id.String()).To(Equal(expected))
			}
		})
	})
})
