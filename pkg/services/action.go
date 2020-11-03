package services

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/guregu/null"
	"go.uber.org/zap"

	"github.com/cmsgov/easi-app/pkg/appcontext"
	"github.com/cmsgov/easi-app/pkg/apperrors"
	"github.com/cmsgov/easi-app/pkg/models"
)

// NewTakeAction is a service to create and execute an action
func NewTakeAction(
	fetch func(context.Context, uuid.UUID) (*models.SystemIntake, error),
	submit func(context.Context, *models.SystemIntake) error,
	reviewNotITRequest func(context.Context, *models.SystemIntake, *models.Action) error,
	reviewReadyForGRT func(context.Context, *models.SystemIntake, *models.Action) error,
	reviewRequestBizCase func(context.Context, *models.SystemIntake, *models.Action) error,
	reviewProvideFeedbackNeedBizCase func(context.Context, *models.SystemIntake, *models.Action) error,
	reviewReadyForGRB func(context.Context, *models.SystemIntake, *models.Action) error,
	issueLCID func(context.Context, *models.SystemIntake, *models.Action) error,
) func(context.Context, *models.Action) error {
	return func(ctx context.Context, action *models.Action) error {
		intake, fetchErr := fetch(ctx, *action.IntakeID)
		if fetchErr != nil {
			return &apperrors.QueryError{
				Err:       fetchErr,
				Operation: apperrors.QueryFetch,
				Model:     intake,
			}
		}

		switch action.ActionType {
		case models.ActionTypeSUBMITINTAKE:
			return submit(ctx, intake)
		case models.ActionTypeNOTITREQUEST:
			return reviewNotITRequest(ctx, intake, action)
		case models.ActionTypeNEEDBIZCASE:
			return reviewRequestBizCase(ctx, intake, action)
		case models.ActionTypeREADYFORGRT:
			return reviewReadyForGRT(ctx, intake, action)
		case models.ActionTypePROVIDEFEEDBACKNEEDBIZCASE:
			return reviewProvideFeedbackNeedBizCase(ctx, intake, action)
		case models.ActionTypeREADYFORGRB:
			return reviewReadyForGRB(ctx, intake, action)
		case models.ActionTypeISSUELCID:
			return issueLCID(ctx, intake, action)
		default:
			return &apperrors.ResourceConflictError{
				Err:        errors.New("invalid action type"),
				Resource:   intake,
				ResourceID: intake.ID.String(),
			}
		}
	}
}

// NewSubmitSystemIntake returns a function that
// executes submit of a system intake
func NewSubmitSystemIntake(
	config Config,
	authorize func(context.Context, *models.SystemIntake) (bool, error),
	update func(context.Context, *models.SystemIntake) (*models.SystemIntake, error),
	validateAndSubmit func(context.Context, *models.SystemIntake) (string, error),
	createAction func(context.Context, *models.Action) (*models.Action, error),
	fetchUserInfo func(context.Context, string) (*models.UserInfo, error),
	emailReviewer func(requester string, intakeID uuid.UUID) error,
) func(context.Context, *models.SystemIntake) error {
	return func(ctx context.Context, intake *models.SystemIntake) error {
		ok, err := authorize(ctx, intake)
		if err != nil {
			return err
		}
		if !ok {
			return &apperrors.UnauthorizedError{Err: err}
		}

		updatedTime := config.clock.Now()
		intake.UpdatedAt = &updatedTime
		intake.Status = models.SystemIntakeStatusINTAKESUBMITTED

		if intake.AlfabetID.Valid {
			err := &apperrors.ResourceConflictError{
				Err:        errors.New("intake has already been submitted to CEDAR"),
				ResourceID: intake.ID.String(),
				Resource:   intake,
			}
			return err
		}

		userInfo, err := fetchUserInfo(ctx, appcontext.Principal(ctx).ID())
		if err != nil {
			return err
		}
		if userInfo == nil || userInfo.Email == "" || userInfo.CommonName == "" || userInfo.EuaUserID == "" {
			return &apperrors.ExternalAPIError{
				Err:       errors.New("user info fetch was not successful"),
				Model:     intake,
				ModelID:   intake.ID.String(),
				Operation: apperrors.Fetch,
				Source:    "CEDAR LDAP",
			}
		}

		intake.SubmittedAt = &updatedTime
		alfabetID, validateAndSubmitErr := validateAndSubmit(ctx, intake)
		if validateAndSubmitErr != nil {
			return validateAndSubmitErr
		}
		if alfabetID == "" {
			return &apperrors.ExternalAPIError{
				Err:       errors.New("submission was not successful"),
				Model:     intake,
				ModelID:   intake.ID.String(),
				Operation: apperrors.Submit,
				Source:    "CEDAR EASi",
			}
		}
		intake.AlfabetID = null.StringFrom(alfabetID)

		action := models.Action{
			IntakeID:       &intake.ID,
			ActionType:     models.ActionTypeSUBMITINTAKE,
			ActorName:      userInfo.CommonName,
			ActorEmail:     userInfo.Email,
			ActorEUAUserID: userInfo.EuaUserID,
		}
		_, err = createAction(ctx, &action)
		if err != nil {
			return &apperrors.QueryError{
				Err:       err,
				Model:     action,
				Operation: apperrors.QueryPost,
			}
		}

		intake, err = update(ctx, intake)
		if err != nil {
			return &apperrors.QueryError{
				Err:       err,
				Model:     intake,
				Operation: apperrors.QuerySave,
			}
		}
		// only send an email when everything went ok
		err = emailReviewer(intake.Requester, intake.ID)
		if err != nil {
			appcontext.ZLogger(ctx).Error("Submit Intake email failed to send: ", zap.Error(err))
		}

		return nil
	}
}

// NewSubmitBusinessCase returns a function that
// executes submit of a business case
func NewSubmitBusinessCase(
	config Config,
	authorize func(context.Context, *models.SystemIntake) (bool, error),
	fetchOpenBusinessCase func(context.Context, uuid.UUID) (*models.BusinessCase, error),
	validateForSubmit func(businessCase *models.BusinessCase) error,
	createAction func(context.Context, *models.Action) (*models.Action, error),
	fetchUserInfo func(context.Context, string) (*models.UserInfo, error),
	updateIntake func(context.Context, *models.SystemIntake) (*models.SystemIntake, error),
	updateBusinessCase func(context.Context, *models.BusinessCase) (*models.BusinessCase, error),
	sendEmail func(requester string, intakeID uuid.UUID) error,
) func(context.Context, *models.SystemIntake) error {
	return func(ctx context.Context, intake *models.SystemIntake) error {
		ok, err := authorize(ctx, intake)
		if err != nil {
			return err
		}
		if !ok {
			return &apperrors.UnauthorizedError{Err: err}
		}

		businessCase, err := fetchOpenBusinessCase(ctx, intake.ID)
		if err != nil {
			return &apperrors.QueryError{
				Err:       err,
				Operation: apperrors.QueryFetch,
				Model:     intake,
			}
		}
		// Uncomment below when UI has changed for unique lifecycle costs
		//err = appvalidation.BusinessCaseForUpdate(businessCase)
		//if err != nil {
		//	return &models.BusinessCase{}, err
		//}
		updatedAt := config.clock.Now()
		businessCase.UpdatedAt = &updatedAt

		if businessCase.InitialSubmittedAt == nil {
			businessCase.InitialSubmittedAt = &updatedAt
		}
		businessCase.LastSubmittedAt = &updatedAt
		err = validateForSubmit(businessCase)
		if err != nil {
			return err
		}

		userInfo, err := fetchUserInfo(ctx, appcontext.Principal(ctx).ID())
		if err != nil {
			return err
		}
		if userInfo == nil || userInfo.Email == "" || userInfo.CommonName == "" || userInfo.EuaUserID == "" {
			return &apperrors.ExternalAPIError{
				Err:       errors.New("user info fetch was not successful"),
				Model:     intake,
				ModelID:   intake.ID.String(),
				Operation: apperrors.Fetch,
				Source:    "CEDAR LDAP",
			}
		}

		action := models.Action{
			IntakeID:       &intake.ID,
			ActionType:     models.ActionTypeSUBMITBIZCASE,
			ActorName:      userInfo.CommonName,
			ActorEmail:     userInfo.Email,
			ActorEUAUserID: userInfo.EuaUserID,
		}
		_, err = createAction(ctx, &action)
		if err != nil {
			return &apperrors.QueryError{
				Err:       err,
				Model:     action,
				Operation: apperrors.QueryPost,
			}
		}

		businessCase, err = updateBusinessCase(ctx, businessCase)
		if err != nil {
			return &apperrors.QueryError{
				Err:       err,
				Model:     businessCase,
				Operation: apperrors.QuerySave,
			}
		}

		intake.Status = models.SystemIntakeStatusBIZCASESUBMITTED
		intake.UpdatedAt = &updatedAt
		intake, err = updateIntake(ctx, intake)
		if err != nil {
			return &apperrors.QueryError{
				Err:       err,
				Model:     intake,
				Operation: apperrors.QuerySave,
			}
		}

		err = sendEmail(businessCase.Requester.String, businessCase.ID)
		if err != nil {
			appcontext.ZLogger(ctx).Error("Submit Business Case email failed to send: ", zap.Error(err))
		}

		return nil
	}
}

// NewTakeActionUpdateStatus returns a function that
// updates the status of a request
func NewTakeActionUpdateStatus(
	config Config,
	newStatus models.SystemIntakeStatus,
	update func(c context.Context, intake *models.SystemIntake) (*models.SystemIntake, error),
	authorize func(context.Context) (bool, error),
	createAction func(context.Context, *models.Action) (*models.Action, error),
	fetchUserInfo func(context.Context, string) (*models.UserInfo, error),
	sendReviewEmail func(emailText string, recipientAddress string) error,
) func(context.Context, *models.SystemIntake, *models.Action) error {
	return func(ctx context.Context, intake *models.SystemIntake, action *models.Action) error {
		ok, err := authorize(ctx)
		if err != nil {
			return err
		}
		if !ok {
			return &apperrors.UnauthorizedError{}
		}

		requesterInfo, err := fetchUserInfo(ctx, intake.EUAUserID)
		if err != nil {
			return err
		}
		if requesterInfo == nil || requesterInfo.Email == "" {
			return &apperrors.ExternalAPIError{
				Err:       errors.New("user info fetch was not successful"),
				Model:     intake,
				ModelID:   intake.ID.String(),
				Operation: apperrors.Fetch,
				Source:    "CEDAR LDAP",
			}
		}

		actorInfo, err := fetchUserInfo(ctx, appcontext.Principal(ctx).ID())
		if err != nil {
			return err
		}
		if actorInfo == nil || actorInfo.Email == "" || actorInfo.CommonName == "" || actorInfo.EuaUserID == "" {
			return &apperrors.ExternalAPIError{
				Err:       errors.New("user info fetch was not successful"),
				Model:     intake,
				ModelID:   intake.ID.String(),
				Operation: apperrors.Fetch,
				Source:    "CEDAR LDAP",
			}
		}

		action.ActorName = actorInfo.CommonName
		action.ActorEmail = actorInfo.Email
		action.ActorEUAUserID = actorInfo.EuaUserID
		_, err = createAction(ctx, action)
		if err != nil {
			return &apperrors.QueryError{
				Err:       err,
				Model:     action,
				Operation: apperrors.QueryPost,
			}
		}

		updatedTime := config.clock.Now()
		intake.UpdatedAt = &updatedTime
		intake.Status = newStatus

		intake, err = update(ctx, intake)
		if err != nil {
			return &apperrors.QueryError{
				Err:       err,
				Model:     intake,
				Operation: apperrors.QuerySave,
			}
		}

		err = sendReviewEmail(intake.GrtReviewEmailBody.String, requesterInfo.Email)
		if err != nil {
			return err
		}

		return nil
	}
}
