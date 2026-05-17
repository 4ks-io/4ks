/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { models_UserEvent } from './models_UserEvent';

export type models_User = {
    createdDate?: string;
    displayName?: string;
    emailAddress?: string;
    events?: Array<models_UserEvent>;
    firstLogin?: boolean;
    id?: string;
    onboardingSource?: string;
    updatedDate?: string;
    username?: string;
    usernameLower?: string;
    welcomeEmailSent?: boolean;
};
