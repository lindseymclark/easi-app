package graph

import (
	"context"
	"fmt"
	"net/url"
	"testing"

	"github.com/99designs/gqlgen/client"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/client/metadata"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/guregu/null"
	_ "github.com/lib/pq" // required for postgres driver in sql
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	ld "gopkg.in/launchdarkly/go-server-sdk.v5"

	"github.com/cmsgov/easi-app/pkg/appconfig"
	"github.com/cmsgov/easi-app/pkg/graph/generated"
	"github.com/cmsgov/easi-app/pkg/models"
	"github.com/cmsgov/easi-app/pkg/storage"
	"github.com/cmsgov/easi-app/pkg/testhelpers"
	"github.com/cmsgov/easi-app/pkg/upload"
)

type GraphQLTestSuite struct {
	suite.Suite
	logger *zap.Logger
	store  *storage.Store
	client *client.Client
}

type mockS3Client struct {
	s3iface.S3API
}

func (m mockS3Client) PutObjectRequest(input *s3.PutObjectInput) (*request.Request, *s3.PutObjectOutput) {

	newRequest := request.New(
		aws.Config{},
		metadata.ClientInfo{},
		request.Handlers{},
		nil,
		&request.Operation{},
		nil,
		nil,
	)

	newRequest.Handlers.Sign.PushFront(func(r *request.Request) {
		r.HTTPRequest.URL = &url.URL{Host: "signed.example.com", Path: "signed/123", Scheme: "https"}
	})
	return newRequest, &s3.PutObjectOutput{}
}

func TestGraphQLTestSuite(t *testing.T) {
	config := testhelpers.NewConfig()

	logger, loggerErr := zap.NewDevelopment()
	if loggerErr != nil {
		panic(loggerErr)
	}

	dbConfig := storage.DBConfig{
		Host:     config.GetString(appconfig.DBHostConfigKey),
		Port:     config.GetString(appconfig.DBPortConfigKey),
		Database: config.GetString(appconfig.DBNameConfigKey),
		Username: config.GetString(appconfig.DBUsernameConfigKey),
		Password: config.GetString(appconfig.DBPasswordConfigKey),
		SSLMode:  config.GetString(appconfig.DBSSLModeConfigKey),
	}

	ldClient, err := ld.MakeCustomClient("fake", ld.Config{Offline: true}, 0)
	assert.NoError(t, err)

	store, err := storage.NewStore(logger, dbConfig, ldClient)
	if err != nil {
		fmt.Printf("Failed to get new database: %v", err)
		t.Fail()
	}

	s3Config := upload.Config{Bucket: "test", Region: "us-west", IsLocal: false}
	mockClient := mockS3Client{}
	s3Client := upload.NewS3ClientUsingClient(mockClient, s3Config)

	schema := generated.NewExecutableSchema(generated.Config{Resolvers: NewResolver(store, ResolverService{}, &s3Client)})
	graphQLClient := client.New(handler.NewDefaultServer(schema))

	storeTestSuite := &GraphQLTestSuite{
		Suite:  suite.Suite{},
		logger: logger,
		store:  store,
		client: graphQLClient,
	}

	suite.Run(t, storeTestSuite)
}

func (s GraphQLTestSuite) TestAccessibilityRequestQuery() {
	ctx := context.Background()

	// ToDo: This whole test would probably be better as an integration test in pkg/integration, so we can see the real
	// functionality and not have to reinitialize all the services here

	intake, intakeErr := s.store.CreateSystemIntake(ctx, &models.SystemIntake{
		ProjectName:            null.StringFrom("Big Project"),
		Status:                 models.SystemIntakeStatusLCIDISSUED,
		RequestType:            models.SystemIntakeRequestTypeNEW,
		BusinessOwner:          null.StringFrom("Firstname Lastname"),
		BusinessOwnerComponent: null.StringFrom("OIT"),
	})
	s.NoError(intakeErr)

	// Can't set a lifecycle ID when creating system intake, so we need to do this to set that
	// so we can then query for the system within the resolver
	lifecycleID, lcidErr := s.store.GenerateLifecycleID(ctx)
	s.NoError(lcidErr)
	intake.LifecycleID = null.StringFrom(lifecycleID)
	_, updateErr := s.store.UpdateSystemIntake(ctx, intake)
	s.NoError(updateErr)

	accessibilityRequest, requestErr := s.store.CreateAccessibilityRequest(ctx, &models.AccessibilityRequest{
		IntakeID: intake.ID,
	})
	s.NoError(requestErr)

	var resp struct {
		AccessibilityRequest struct {
			ID     string
			System struct {
				ID            string
				Name          string
				LCID          string
				BusinessOwner struct {
					Name      string
					Component string
				}
			}
		}
	}

	// TODO we're supposed to be able to pass variables as additional arguments using client.Var()
	// but it wasn't working for me.
	s.client.MustPost(fmt.Sprintf(
		`query {
			accessibilityRequest(id: "%s") {
				id
				system {
					id
					name
					lcid
					businessOwner {
						name
						component
					}
				}
			}
		}`, accessibilityRequest.ID), &resp)

	s.Equal(accessibilityRequest.ID.String(), resp.AccessibilityRequest.ID)
	s.Equal(intake.ID.String(), resp.AccessibilityRequest.System.ID)
	s.Equal("Big Project", resp.AccessibilityRequest.System.Name)
	s.Equal(lifecycleID, resp.AccessibilityRequest.System.LCID)
	s.Equal("Firstname Lastname", resp.AccessibilityRequest.System.BusinessOwner.Name)
	s.Equal("OIT", resp.AccessibilityRequest.System.BusinessOwner.Component)
}

func (s GraphQLTestSuite) TestGeneratePresignedUploadURLMutation() {
	var resp struct {
		GeneratePresignedUploadURL struct {
			URL        string
			UserErrors []struct {
				Message string
				Path    []string
			}
		}
	}

	s.client.MustPost(
		`mutation {
			generatePresignedUploadURL(input: {mimeType: "application/pdf"}) {
				url
				userErrors {
					message
					path
				}
			}
		}`, &resp)

	s.Equal("https://signed.example.com/signed/123", resp.GeneratePresignedUploadURL.URL)
	s.Equal(0, len(resp.GeneratePresignedUploadURL.UserErrors))
}
