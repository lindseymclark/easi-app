import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link, useHistory, useParams } from 'react-router-dom';
import { useMutation, useQuery } from '@apollo/client';
import { Button } from '@trussworks/react-uswds';
import { Field, Form, Formik, FormikProps } from 'formik';
import { DateTime } from 'luxon';
import CreateTestDateQuery from 'queries/CreateTestDateQuery';
import GetAccessibilityRequestQuery from 'queries/GetAccessibilityRequestQuery';
import { CreateTestDate } from 'queries/types/CreateTestDate';
import { GetAccessibilityRequest } from 'queries/types/GetAccessibilityRequest';

import Footer from 'components/Footer';
import Header from 'components/Header';
import MainContent from 'components/MainContent';
import PageWrapper from 'components/PageWrapper';
import {
  DateInputDay,
  DateInputMonth,
  DateInputYear
} from 'components/shared/DateInput';
import { ErrorAlert, ErrorAlertMessage } from 'components/shared/ErrorAlert';
import FieldErrorMsg from 'components/shared/FieldErrorMsg';
import FieldGroup from 'components/shared/FieldGroup';
import HelpText from 'components/shared/HelpText';
import Label from 'components/shared/Label';
import { RadioField } from 'components/shared/RadioField';
import { NavLink, SecondaryNav } from 'components/shared/SecondaryNav';
import TextField from 'components/shared/TextField';
import { TestDateForm } from 'types/accessibilityRequest';
import flattenErrors from 'utils/flattenErrors';
import { TestDateValidationSchema } from 'validations/testDateSchema';

const TestDate = () => {
  const { t } = useTranslation('accessibility');
  const { accessibilityRequestId } = useParams<{
    accessibilityRequestId: string;
  }>();
  const [mutate, mutationResult] = useMutation<CreateTestDate>(
    CreateTestDateQuery,
    {
      errorPolicy: 'all'
    }
  );
  const { data } = useQuery<GetAccessibilityRequest>(
    GetAccessibilityRequestQuery,
    {
      variables: {
        id: accessibilityRequestId
      }
    }
  );
  const history = useHistory();
  const initialValues: TestDateForm = {
    testType: null,
    dateMonth: '',
    dateDay: '',
    dateYear: '',
    score: {
      isPresent: false,
      value: ''
    }
  };

  const onSubmit = (values: TestDateForm) => {
    const testDate = DateTime.fromObject({
      day: Number(values.dateDay),
      month: Number(values.dateMonth),
      year: Number(values.dateYear)
    });

    mutate({
      variables: {
        input: {
          date: testDate,
          score: values.score.isPresent
            ? Math.round(parseFloat(values.score.value) * 10)
            : null,
          testType: values.testType,
          requestID: accessibilityRequestId
        }
      }
    }).then(() => {
      history.push(`/508/requests/${accessibilityRequestId}`, {
        confirmationText: t('createTestDate.confirmation', {
          date: testDate.toLocaleString(DateTime.DATE_FULL),
          requestName: data?.accessibilityRequest?.system?.name
        })
      });
    });
  };

  return (
    <PageWrapper className="add-test-date">
      <Header />
      <MainContent className="margin-bottom-5">
        <SecondaryNav>
          <NavLink to={`/508/requests/${accessibilityRequestId}`}>
            {t('tabs.accessibilityRequests')}
          </NavLink>
        </SecondaryNav>
        <div className="grid-container padding-y-6">
          <Formik
            initialValues={initialValues}
            onSubmit={onSubmit}
            validationSchema={TestDateValidationSchema}
            validateOnBlur={false}
            validateOnChange={false}
            validateOnMount={false}
          >
            {(formikProps: FormikProps<TestDateForm>) => {
              const { errors, setFieldValue, values } = formikProps;
              const flatErrors = flattenErrors(errors);
              return (
                <>
                  {Object.keys(errors).length > 0 && (
                    <ErrorAlert
                      testId="test-date-errors"
                      classNames="margin-bottom-4"
                      heading="Please check and fix the following"
                    >
                      {Object.keys(flatErrors).map(key => {
                        return (
                          <ErrorAlertMessage
                            key={`Error.${key}`}
                            errorKey={key}
                            message={flatErrors[key]}
                          />
                        );
                      })}
                    </ErrorAlert>
                  )}
                  {mutationResult.error && (
                    <ErrorAlert heading="System error">
                      <ErrorAlertMessage
                        message={mutationResult.error.message}
                        errorKey="system"
                      />
                    </ErrorAlert>
                  )}
                  <h1 className="margin-bottom-1">
                    {t('createTestDate.addTestDateHeader', {
                      requestName: data?.accessibilityRequest?.system?.name
                    })}
                  </h1>
                  <div className="grid-row grid-gap-lg">
                    <div className="grid-col-9">
                      <Form>
                        <FieldGroup error={!!flatErrors.testType}>
                          <fieldset className="usa-fieldset">
                            <legend className="usa-label margin-bottom-1">
                              {t('createTestDate.testTypeHeader')}
                            </legend>
                            <FieldErrorMsg>{flatErrors.testType}</FieldErrorMsg>

                            <Field
                              as={RadioField}
                              checked={values.testType === 'INITIAL'}
                              id="TestDate-TestTypeInitial"
                              name="testType"
                              label="Initial"
                              value="INITIAL"
                            />
                            <Field
                              as={RadioField}
                              checked={values.testType === 'REMEDIATION'}
                              id="TestDate-TestTypeRemediation"
                              name="testType"
                              label="Remediation"
                              value="REMEDIATION"
                            />
                          </fieldset>
                        </FieldGroup>
                        {/* GRT Date Fields */}
                        <FieldGroup
                          error={
                            !!flatErrors.dateMonth ||
                            !!flatErrors.dateDay ||
                            !!flatErrors.dateYear ||
                            !!flatErrors.validDate
                          }
                        >
                          <fieldset className="usa-fieldset margin-top-4">
                            <legend className="usa-label margin-bottom-1">
                              {t('createTestDate.dateHeader')}
                            </legend>
                            <HelpText
                              id="TestDate-DateHelp"
                              className="margin-bottom-2"
                            >
                              {t('createTestDate.dateHelpText')}
                            </HelpText>
                            <FieldErrorMsg>
                              {flatErrors.dateMonth}
                            </FieldErrorMsg>
                            <FieldErrorMsg>{flatErrors.dateDay}</FieldErrorMsg>
                            <FieldErrorMsg>{flatErrors.dateYear}</FieldErrorMsg>
                            <FieldErrorMsg>
                              {flatErrors.validDate}
                            </FieldErrorMsg>
                            <div
                              className="usa-memorable-date"
                              style={{ marginTop: '-2rem' }}
                            >
                              <div className="usa-form-group usa-form-group--month">
                                <Label
                                  htmlFor="TestDate-DateMonth"
                                  className="text-normal"
                                >
                                  {t('general:date.month')}
                                </Label>
                                <Field
                                  as={DateInputMonth}
                                  error={!!flatErrors.dateMonth}
                                  id="TestDate-DateMonth"
                                  name="dateMonth"
                                />
                              </div>
                              <div className="usa-form-group usa-form-group--day">
                                <Label
                                  htmlFor="TestDate-DateDay"
                                  className="text-normal"
                                >
                                  {t('general:date.day')}
                                </Label>
                                <Field
                                  as={DateInputDay}
                                  error={!!flatErrors.dateDay}
                                  id="TestDate-DateDay"
                                  name="dateDay"
                                />
                              </div>
                              <div className="usa-form-group usa-form-group--year">
                                <Label
                                  htmlFor="TestDate-DateYear"
                                  className="text-normal"
                                >
                                  {t('general:date.year')}
                                </Label>
                                <Field
                                  as={DateInputYear}
                                  error={!!flatErrors.dateYear}
                                  id="TestDate-DateYear"
                                  name="dateYear"
                                />
                              </div>
                            </div>
                          </fieldset>
                        </FieldGroup>
                        <FieldGroup
                          scrollElement="score.isPresent"
                          error={!!flatErrors['score.isPresent']}
                        >
                          <fieldset className="usa-fieldset margin-top-4">
                            <legend className="usa-label margin-bottom-1">
                              {t('createTestDate.scoreHeader')}
                            </legend>

                            <FieldErrorMsg>
                              {flatErrors['score.isPresent']}
                            </FieldErrorMsg>
                            <Field
                              as={RadioField}
                              checked={values.score.isPresent === false}
                              id="TestDate-HasScoreNo"
                              name="score.isPresent"
                              label="No"
                              onChange={() => {
                                setFieldValue('score.isPresent', false);
                                setFieldValue('score.name', '');
                              }}
                              value={false}
                            />
                            <Field
                              as={RadioField}
                              checked={values.score.isPresent === true}
                              id="TestDate-HasScoreYes"
                              name="score.isPresent"
                              label="Yes"
                              onChange={() => {
                                setFieldValue('score.isPresent', true);
                              }}
                              value
                              aria-describedby="TestDate-ScoreYes"
                            />
                            {values.score.isPresent && (
                              <div className="width-card-lg margin-top-neg-2 margin-left-4 margin-bottom-1">
                                <FieldGroup
                                  scrollElement="score.value"
                                  error={!!flatErrors['score.value']}
                                >
                                  <Label htmlFor="TestDate-ScoreValue">
                                    {t('createTestDate.scoreValueHeader')}
                                  </Label>
                                  <Label
                                    htmlFor="TestDate-ScoreValue"
                                    className="usa-sr-only"
                                  >
                                    {t('createTestDate.scoreValueSRHelpText')}
                                  </Label>
                                  <FieldErrorMsg>
                                    {flatErrors['score.value']}
                                  </FieldErrorMsg>
                                  <div className="display-flex">
                                    <div className="width-10">
                                      <Field
                                        as={TextField}
                                        error={!!flatErrors['score.value']}
                                        id="TestDate-ScoreValue"
                                        maxLength={4}
                                        name="score.value"
                                      />
                                    </div>
                                    <div className="bg-black text-white width-5 margin-top-05 display-flex flex-justify-center flex-align-center">
                                      <span className="text-bold">%</span>
                                    </div>
                                  </div>
                                </FieldGroup>
                              </div>
                            )}
                          </fieldset>
                        </FieldGroup>
                        <Button
                          className="margin-top-2 display-block"
                          type="submit"
                        >
                          {t('createTestDate.submitButton')}
                        </Button>
                        <Link
                          to={`/508/requests/${accessibilityRequestId}`}
                          className="margin-top-2 display-block"
                        >
                          {t('createTestDate.cancel')}
                        </Link>
                      </Form>
                    </div>
                  </div>
                </>
              );
            }}
          </Formik>
        </div>
      </MainContent>
      <Footer />
    </PageWrapper>
  );
};

export default TestDate;
