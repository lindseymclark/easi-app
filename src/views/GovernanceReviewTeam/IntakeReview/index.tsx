import React from 'react';
import { useTranslation } from 'react-i18next';
import { Link } from 'react-router-dom';
import { Link as UswdsLink } from '@trussworks/react-uswds';

import SystemIntakeReview from 'components/SystemIntakeReview';
import { SystemIntakeForm } from 'types/systemIntake';

type IntakeReviewProps = {
  systemIntake: SystemIntakeForm;
};

const IntakeReview = ({ systemIntake }: IntakeReviewProps) => {
  const { t } = useTranslation('governanceReviewTeam');

  return (
    <div>
      <h1 className="margin-top-0">{t('general:intake')}</h1>
      <SystemIntakeReview systemIntake={systemIntake} />
      <UswdsLink
        className="usa-button margin-top-5"
        variant="unstyled"
        to={`/governance-review-team/${systemIntake.id}/actions`}
        asCustom={Link}
      >
        Take an action
      </UswdsLink>
    </div>
  );
};

export default IntakeReview;