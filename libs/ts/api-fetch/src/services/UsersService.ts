/* istanbul ignore file */
/* tslint:disable */
/* eslint-disable */
import type { dtos_CreateUser } from '../models/dtos_CreateUser';
import type { dtos_KitchenPassResponse } from '../models/dtos_KitchenPassResponse';
import type { dtos_TestUsernameRequest } from '../models/dtos_TestUsernameRequest';
import type { dtos_TestUsernameResponse } from '../models/dtos_TestUsernameResponse';
import type { dtos_UpdateUser } from '../models/dtos_UpdateUser';
import type { models_User } from '../models/models_User';

import type { CancelablePromise } from '../core/CancelablePromise';
import type { BaseHttpRequest } from '../core/BaseHttpRequest';

export class UsersService {

    constructor(public readonly httpRequest: BaseHttpRequest) {}

    /**
     * Create a new User
     * Create a new User
     * @param user User Data
     * @returns models_User OK
     * @throws ApiError
     */
    public postApiUser(
        user: dtos_CreateUser,
    ): CancelablePromise<models_User> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/api/user',
            body: user,
        });
    }

    /**
     * Get Authenticated User
     * Get Authenticated User
     * @returns models_User OK
     * @throws ApiError
     */
    public getApiUser(): CancelablePromise<models_User> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/api/user/',
        });
    }

    /**
     * Head Authenticated user
     * Head Authenticated user
     * @returns any OK
     * @throws ApiError
     */
    public headApiUser(): CancelablePromise<any> {
        return this.httpRequest.request({
            method: 'HEAD',
            url: '/api/user/',
            errors: {
                400: `Invalid Request`,
                404: `Record Not Found`,
                500: `Internal Error`,
            },
        });
    }

    /**
     * Update User
     * Update User
     * @param payload User Data
     * @returns models_User OK
     * @throws ApiError
     */
    public patchApiUser(
        payload: dtos_UpdateUser,
    ): CancelablePromise<models_User> {
        return this.httpRequest.request({
            method: 'PATCH',
            url: '/api/user/',
            body: payload,
        });
    }

    /**
     * Delete User
     * Delete User
     * @param id Event ID
     * @returns any OK
     * @throws ApiError
     */
    public deleteApiUserEvents(
        id: string,
    ): CancelablePromise<any> {
        return this.httpRequest.request({
            method: 'DELETE',
            url: '/api/user/events/{id}',
            path: {
                'id': id,
            },
        });
    }

    /**
     * Get AI Kitchen Pass status
     * Returns whether the authenticated user has an active AI Kitchen Pass plus the current skill URL and copy text when enabled.
     * @returns dtos_KitchenPassResponse OK
     * @throws ApiError
     */
    public getApiUserKitchenPass(): CancelablePromise<dtos_KitchenPassResponse> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/api/user/kitchen-pass',
        });
    }

    /**
     * Create or rotate AI Kitchen Pass
     * Creates an AI Kitchen Pass for the authenticated user, or rotates the existing pass and immediately invalidates the previous token.
     * @returns dtos_KitchenPassResponse OK
     * @throws ApiError
     */
    public postApiUserKitchenPass(): CancelablePromise<dtos_KitchenPassResponse> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/api/user/kitchen-pass',
        });
    }

    /**
     * Revoke AI Kitchen Pass
     * Revokes the authenticated user's AI Kitchen Pass and immediately disables further bearer-token access for that pass.
     * @returns dtos_KitchenPassResponse OK
     * @throws ApiError
     */
    public deleteApiUserKitchenPass(): CancelablePromise<dtos_KitchenPassResponse> {
        return this.httpRequest.request({
            method: 'DELETE',
            url: '/api/user/kitchen-pass',
        });
    }

    /**
     * Get All Users
     * Get All Users
     * @returns models_User OK
     * @throws ApiError
     */
    public getApiUsers(): CancelablePromise<Array<models_User>> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/api/users/',
        });
    }

    /**
     * Returns username validity and availability
     * Returns username validity and availability
     * @param username Username Data
     * @returns dtos_TestUsernameResponse OK
     * @throws ApiError
     */
    public postApiUsersUsername(
        username: dtos_TestUsernameRequest,
    ): CancelablePromise<dtos_TestUsernameResponse> {
        return this.httpRequest.request({
            method: 'POST',
            url: '/api/users/username',
            body: username,
        });
    }

    /**
     * Get a User (by ID)
     * Get a User (by ID)
     * @param id User ID
     * @returns models_User OK
     * @throws ApiError
     */
    public getApiUsers1(
        id: string,
    ): CancelablePromise<models_User> {
        return this.httpRequest.request({
            method: 'GET',
            url: '/api/users/{id}',
            path: {
                'id': id,
            },
        });
    }

    /**
     * Delete User
     * Delete User
     * @param userId User ID
     * @returns any OK
     * @throws ApiError
     */
    public deleteApiUsers(
        userId: string,
    ): CancelablePromise<any> {
        return this.httpRequest.request({
            method: 'DELETE',
            url: '/api/users/{userID}',
            path: {
                'userID': userId,
            },
        });
    }

}