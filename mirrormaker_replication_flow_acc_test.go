package aiven

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"math/rand"
	"os"
	"strconv"
)

var _ = Describe("MirrorMaker 2 Replication flow", func() {
	var (
		projectName string
		project     *Project
		err         error
	)

	BeforeEach(func() {
		projectName = "test-acc-pr-mm-" + strconv.Itoa(rand.Int())
		project, err = client.Projects.Create(CreateProjectRequest{
			Project: projectName,
			CardID:  os.Getenv("AIVEN_CARD_ID"),
		})
	})

	Context("Create new project", func() {
		It("should not error", func() {
			Expect(err).NotTo(HaveOccurred())
		})

		It("should populate fields properly", func() {
			Expect(project).NotTo(BeNil())

			if project != nil {
				Expect(project.Name).NotTo(BeEmpty())
				Expect(project.AccountId).To(BeEmpty())
			}
		})

		// mirror maker service
		var (
			serviceName string
			service     *Service
			errS        error
		)

		JustBeforeEach(func() {
			serviceName = "test-acc-sr-mm-" + strconv.Itoa(rand.Int())
			service, errS = client.Services.Create(projectName, CreateServiceRequest{
				Cloud:        "google-europe-west1",
				Plan:         "startup-4",
				ProjectVPCID: nil,
				ServiceName:  serviceName,
				ServiceType:  "kafka_mirrormaker",
			})
		})

		Context("Create new mirror maker service", func() {
			It("should not error", func() {
				Expect(errS).NotTo(HaveOccurred())
			})

			It("should populate fields properly", func() {
				Expect(service).NotTo(BeNil())

				if service != nil {
					Expect(service.Name).NotTo(BeEmpty())
					Expect(service.Plan).NotTo(BeEmpty())
					Expect(service.Type).Should(Equal("kafka_mirrormaker"))

					Eventually(func() string {
						service, _ = client.Services.Get(projectName, serviceName)
						return service.State
					}).ShouldNot(Equal("RUNNING"))
				}
			})

			// mirror maker replication flow
			var (
				errR error
			)

			JustBeforeEach(func() {
				errR = client.KafkaMirrorMakerReplicationFlow.Create(projectName, serviceName, MirrorMakerReplicationFlowRequest{
					ReplicationFlow{
						Enabled:         false,
						SourceCluster:   "source",
						TargetCluster:   "target",
						Topics:          []string{".*"},
						TopicsBlacklist: []string{},
					},
				})
			})

			Context("Create new mirror maker replication flow", func() {
				It("should not error", func() {
					Expect(errR).NotTo(HaveOccurred())
				})

				It("should populate fields properly", func() {
					r, errG := client.KafkaMirrorMakerReplicationFlow.Get(projectName, serviceName, "source", "target")
					Expect(errG).NotTo(HaveOccurred())
					Expect(r).NotTo(BeNil())

					if r != nil {
						Expect(r.ReplicationFlow.TargetCluster).NotTo(BeEmpty())
						Expect(r.ReplicationFlow.SourceCluster).NotTo(BeEmpty())
					}
				})

				It("should update", func() {
					r, errU := client.KafkaMirrorMakerReplicationFlow.Update(projectName, serviceName, "source", "target", MirrorMakerReplicationFlowRequest{
						ReplicationFlow: ReplicationFlow{
							Enabled:         false,
							Topics:          []string{".*"},
							TopicsBlacklist: []string{"test"},
						},
					})
					Expect(errU).NotTo(HaveOccurred())
					Expect(r).NotTo(BeNil())

					if r != nil {
						Expect(r.ReplicationFlow.TargetCluster).NotTo(BeEmpty())
						Expect(r.ReplicationFlow.SourceCluster).NotTo(BeEmpty())
						Expect(len(r.ReplicationFlow.TopicsBlacklist)).Should(Equal(1))
					}
				})

				AfterEach(func() {
					if errD := client.KafkaMirrorMakerReplicationFlow.Delete(projectName, serviceName, "source", "target"); errD != nil {
						Fail("cannot delete mirror maker replication flow:" + errD.Error())
					}
				})
			})
		})

		AfterEach(func() {
			if errD := client.Services.Delete(projectName, serviceName); errD != nil {
				Fail("cannot delete service:" + errD.Error())
			}
		})
	})

	AfterEach(func() {
		err = client.Projects.Delete(projectName)
		if err != nil {
			Fail("cannot delete project : " + err.Error())
		}
	})
})