/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */

import type { models_MediaBestUse } from './models_MediaBestUse';
import type { models_MediaSource } from './models_MediaSource';
import type { models_MediaStatus } from './models_MediaStatus';
import type { models_RecipeMediaVariant } from './models_RecipeMediaVariant';

export type models_RecipeMedia = {
    bestUse?: models_MediaBestUse;
    contentType?: string;
    createdDate?: string;
    id?: string;
    ownerId?: string;
    recipeId?: string;
    rootRecipeId?: string;
    source?: models_MediaSource;
    status?: models_MediaStatus;
    updatedDate?: string;
    variants?: Array<models_RecipeMediaVariant>;
};
